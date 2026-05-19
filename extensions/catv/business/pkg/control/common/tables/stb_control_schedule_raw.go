package tables

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	batchV1 "k8s.io/api/batch/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
	control_common "ntels.com/pharos/extensions/catv/business/pkg/control/common"
	"ntels.com/pharos/extensions/catv/business/pkg/control/common/kubernetes"
)

const (
	ScheduleTypeImmediately = "immediately"
	ScheduleTypeOnce        = "once"
	ScheduleTypeRepeat      = "repeat"
)

type StbControlScheduleRaw struct {
	ID   string `json:"id" db:"id"`
	Name string `json:"name" db:"name"`

	ScheduleType       string        `json:"schedule_type" db:"schedule_type"`
	ScheduleSpecOnce   *orm.Datetime `json:"schedule_spec_once,omitempty" db:"schedule_spec_once"`
	ScheduleSpecRepeat *string       `json:"schedule_spec_repeat,omitempty" db:"schedule_spec_repeat"`

	Sos         []string `json:"sos" db:"sos"`
	L3s         []string `json:"l3s" db:"l3s"`
	Cells       []string `json:"cells" db:"cells"`
	Settopboxes []string `json:"settopboxes" db:"settopboxes"`

	WorkType  string  `json:"work_type" db:"work_type"`
	WorkValue *string `json:"work_value,omitempty" db:"work_value"`

	AreaType string   `json:"area_type" db:"area_type"`
	AreaIDs  []string `json:"area_ids" db:"area_ids"`

	CreateTime orm.Datetime `json:"create_time" db:"create_time"`
	UpdateTime orm.Datetime `json:"update_time" db:"update_time"`
	IsDeleted  uint8        `json:"is_deleted" db:"is_deleted"`
}

func (r StbControlScheduleRaw) Validate() error {
	switch r.ScheduleType {
	case ScheduleTypeImmediately:
	case ScheduleTypeOnce:
		if r.ScheduleSpecOnce == nil {
			return fmt.Errorf("once_spec is required when run_type is 'once'")
		}
		if r.ScheduleSpecOnce.Time.Before(time.Now().UTC()) {
			return fmt.Errorf("once_spec must be a future time")
		}
	case ScheduleTypeRepeat:
		if r.ScheduleSpecRepeat == nil {
			return fmt.Errorf("repeat_spec is required when run_type is 'repeat'")
		}

		parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
		if _, err := parser.Parse(*r.ScheduleSpecRepeat); err != nil {
			return fmt.Errorf("invalid repeat_spec cron expression: %w", err)
		}
	default:
		return fmt.Errorf("invalid schedule_type: must be one of immediately, once, repeat")
	}

	// AreaIDs is required for non-"all" types
	if strings.ToLower(r.AreaType) != AreaTypeAll && len(r.AreaIDs) == 0 {
		return fmt.Errorf("area_ids is required when area_type is not 'all'")
	}

	// Validate work_type (should match valid work types from command package)
	if r.WorkType == "" {
		return fmt.Errorf("work_type is required")
	}

	return nil
}

func (r StbControlScheduleRaw) Execute(configPath string, config common.Config, requestTime orm.Datetime) {
	k8sJobNames := r.makeK8sJobNames(config)

	results, ok := r.makeResults(config, requestTime, k8sJobNames)
	r.insertResults(config, results, "", "", "")
	if !ok {
		slog.Info("Failed to make control job results, skipping Kubernetes Job creation", "id", r.ID, "name", r.Name)
		return
	}

	time.Sleep(1 * time.Second)
	for _, k8sJobName := range k8sJobNames {
		if err := r.createK8sJob(configPath, config, k8sJobName); err != nil {
			slog.Error("Failed to create Kubernetes Job for control schedule execution", "id", r.ID, "name", r.Name, "k8s_job_name", k8sJobName, "error", err)
			r.insertResults(config, results, k8sJobName, control_common.ResponseCodeError, fmt.Sprintf("failed to create Kubernetes Job: %v", err))
			continue
		}
	}
}

func (r StbControlScheduleRaw) makeResults(config common.Config, requestTime orm.Datetime, k8sJobNames []string) ([]StbControlScheduleResultRaw, bool) {
	baseResult := StbControlScheduleResultRaw{
		ID: fmt.Sprintf("%s-%s", r.ID, requestTime.Time.Format("20060102150405")),

		RequestTime: requestTime,

		ScheduleID:   r.ID,
		ScheduleName: r.Name,
		ScheduleType: r.ScheduleType,

		K8sJobName: "",

		ProgressStatus: ProgressStatusPending,
		WorkType:       r.WorkType,
		WorkValue:      r.WorkValue,
	}

	stbInformationRaws, err := NewStbInformationTable(config).GetStbInformationRaws(r.AreaType, r.AreaIDs)
	if err != nil {
		slog.Error("Failed to get device information for control schedule execution", "area_type", r.AreaType, "area_ids", r.AreaIDs, "error", err)

		baseResult.ResultCode = control_common.ResponseCodeError
		baseResult.ResultMessage = fmt.Sprintf("failed to get device information: %v", err)

		return []StbControlScheduleResultRaw{baseResult}, false
	}
	if len(stbInformationRaws) == 0 {
		slog.Error("No devices found for the specified area", "area_type", r.AreaType, "area_ids", r.AreaIDs)

		baseResult.ResultCode = control_common.ResponseCodeError
		baseResult.ResultMessage = "no devices found for the specified area"

		return []StbControlScheduleResultRaw{baseResult}, false
	}

	var results []StbControlScheduleResultRaw

	for index, raw := range stbInformationRaws {
		baseResult.K8sJobName = k8sJobNames[index%len(k8sJobNames)]

		baseResult.StbMdlNm = raw.StbMdlNm
		baseResult.CmMacAddr = raw.CmMacAddr
		baseResult.CmIpAddr = raw.CmIpAddr
		baseResult.StbMacAddr = raw.StbMacAddr
		baseResult.SrcIpAddr = raw.SrcIpAddr

		results = append(results, baseResult)
	}

	return results, true
}

func (r StbControlScheduleRaw) insertResults(config common.Config, results []StbControlScheduleResultRaw, whereK8sJobName string, resultCode, resultMessage string) {
	var updatedResults []StbControlScheduleResultRaw

	for index := range results {
		if results[index].K8sJobName == whereK8sJobName {
			if resultCode != "" {
				results[index].ResultCode = resultCode
			}
			if resultMessage != "" {
				results[index].ResultMessage = resultMessage
			}
		}

		updatedResults = append(updatedResults, results[index])
	}

	stbControlScheduleResultTable := NewStbControlScheduleResultTable(config)
	if err := stbControlScheduleResultTable.Inserts(updatedResults); err != nil {
		slog.Error("Failed to insert control schedule results into database", "id", r.ID, "name", r.Name, "error", err)
	}
}

func (r StbControlScheduleRaw) makeK8sJobNames(config common.Config) []string {
	const prefix = "catv-control"

	var k8sJobNames []string

	for i := 0; i < config.Catv.Control.K8sJob.Count; i++ {
		k8sJobName := fmt.Sprintf("%s-%s-%d", prefix, uuid.New().String(), i)
		if errs := validation.IsDNS1123Label(k8sJobName); len(errs) > 0 {
			var newK8sJobName = fmt.Sprintf("%s-%s-%d", prefix, "name-check-needed", i)
			slog.Warn("Generated Kubernetes Job name does not conform to DNS-1123 label format", "original_name", k8sJobName, "new_name", newK8sJobName, "errors", errs)
			k8sJobName = newK8sJobName
		}

		k8sJobNames = append(k8sJobNames, k8sJobName)
	}

	return k8sJobNames
}

func (r StbControlScheduleRaw) createK8sJob(configPath string, config common.Config, k8sJobName string) error {
	const labelKey = "app"
	const labelValue = "catv-control-job"

	namespace := config.Catv.Control.K8sJob.Namespace
	image := config.Catv.Control.K8sJob.Image
	serviceAccountName := config.Catv.Control.K8sJob.ServiceAccountName

	envMap := make(map[string]string)
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 && slices.Contains(config.Catv.Control.K8sJob.AllowedEnvVars, parts[0]) {
			envMap[parts[0]] = parts[1]
		}
	}

	var env []coreV1.EnvVar
	for envName, envValue := range envMap {
		env = append(env, coreV1.EnvVar{
			Name:  envName,
			Value: envValue,
		})
	}

	job := batchV1.Job{
		TypeMeta: metaV1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      k8sJobName,
			Namespace: namespace,
			Labels: map[string]string{
				labelKey: labelValue,
			},
		},
		Spec: batchV1.JobSpec{
			ActiveDeadlineSeconds:   &config.Catv.Control.K8sJob.ActiveDeadlineSeconds,
			BackoffLimit:            &config.Catv.Control.K8sJob.BackoffLimit,
			TTLSecondsAfterFinished: &config.Catv.Control.K8sJob.TTLSecondsAfterFinished,
			Template: coreV1.PodTemplateSpec{
				ObjectMeta: metaV1.ObjectMeta{
					Labels: map[string]string{
						labelKey: labelValue,
					},
				},
				Spec: coreV1.PodSpec{
					Volumes: []coreV1.Volume{
						{
							Name: config.Catv.Control.K8sJob.ConfigMapName,
							VolumeSource: coreV1.VolumeSource{
								ConfigMap: &coreV1.ConfigMapVolumeSource{
									LocalObjectReference: coreV1.LocalObjectReference{
										Name: config.Catv.Control.K8sJob.ConfigMapName,
									},
								},
							},
						},
						{
							Name: config.Catv.Control.K8sJob.SecretName,
							VolumeSource: coreV1.VolumeSource{
								Secret: &coreV1.SecretVolumeSource{
									SecretName: config.Catv.Control.K8sJob.SecretName,
								},
							},
						},
					},
					Containers: []coreV1.Container{
						{
							Name:            k8sJobName,
							Image:           image,
							Command:         []string{"/opt/pharos/bin/pharos", control_common.CommandUse, "--config=" + configPath, "--k8s-job-name=" + k8sJobName},
							Env:             env,
							ImagePullPolicy: coreV1.PullAlways,
							VolumeMounts: []coreV1.VolumeMount{
								{
									Name:      config.Catv.Control.K8sJob.ConfigMapName,
									MountPath: filepath.Dir(configPath),
								},
							},
						},
					},
					Affinity: &coreV1.Affinity{
						PodAntiAffinity: &coreV1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []coreV1.PodAffinityTerm{
								{
									LabelSelector: &metaV1.LabelSelector{
										MatchExpressions: []metaV1.LabelSelectorRequirement{
											{
												Key:      labelKey,
												Operator: metaV1.LabelSelectorOpIn,
												Values:   []string{labelValue},
											},
										},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
							PreferredDuringSchedulingIgnoredDuringExecution: []coreV1.WeightedPodAffinityTerm{
								{
									Weight: 100,
									PodAffinityTerm: coreV1.PodAffinityTerm{
										LabelSelector: &metaV1.LabelSelector{
											MatchExpressions: []metaV1.LabelSelectorRequirement{
												{
													Key:      labelKey,
													Operator: metaV1.LabelSelectorOpIn,
													Values:   []string{labelValue},
												},
											},
										},
										TopologyKey: "kubernetes.io/hostname",
									},
								},
							},
						},
					},
					RestartPolicy:      coreV1.RestartPolicyNever,
					ServiceAccountName: serviceAccountName,
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return kubernetes.Post(ctx, "batch", "v1", namespace, "jobs", &job)
}
