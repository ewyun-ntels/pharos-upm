package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/IBM/sarama"
)

var (
	brokers         = flag.String("brokers", "adamscatvmq01:29091,adamscatvmq02:29091,adamscatvmq03:29091,adamscatvmq04:29091,adamscatvmq05:29091", "Kafka broker addresses (comma-separated)")
	consumerGroup   = flag.String("group", "", "Consumer group name (empty for all groups)")
	listGroups      = flag.Bool("list", false, "List all consumer groups")
	watch           = flag.Bool("watch", false, "Watch mode - refresh every 5 seconds")
	detail          = flag.Bool("detail", false, "Show detailed member information")
	web             = flag.Bool("web", false, "Start web UI server")
	webPort         = flag.String("port", "8080", "Web UI port")
	refreshInterval = flag.Int("refresh", 5, "Auto-refresh interval in seconds for web UI (default: 5)")
)

type ConsumerGroupInfo struct {
	Group           string
	Topic           string
	Partition       int32
	CurrentOffset   int64
	LogEndOffset    int64
	Lag             int64
	ClientID        string
	Host            string
	MemberID        string
	GroupInstanceID string
}

func main() {
	flag.Parse()

	brokerList := strings.Split(*brokers, ",")

	config := sarama.NewConfig()
	config.Version = sarama.V2_6_0_0
	config.Consumer.Return.Errors = true

	admin, err := sarama.NewClusterAdmin(brokerList, config)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error creating cluster admin: %v\n", err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
	defer func() { _ = admin.Close() }()

	client, err := sarama.NewClient(brokerList, config)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = client.Close() }()

	coordinator, err := sarama.NewConsumerGroupFromClient("monitor", client)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error creating coordinator: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = coordinator.Close() }()

	if *web {
		startWebServer(admin, client)
		return
	}

	if *listGroups {
		if err := listConsumerGroups(admin); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error listing groups: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *detail {
		if err := showMemberDetailInfo(admin, *consumerGroup); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *watch {
		for {
			clearScreen()
			if err := showConsumerGroupInfo(admin, client, *consumerGroup); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			time.Sleep(5 * time.Second)
		}
	} else {
		if err := showConsumerGroupInfo(admin, client, *consumerGroup); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func listConsumerGroups(admin sarama.ClusterAdmin) error {
	groups, err := admin.ListConsumerGroups()
	if err != nil {
		return fmt.Errorf("failed to list consumer groups: %w", err)
	}

	fmt.Printf("Found %d consumer groups:\n\n", len(groups))

	groupNames := make([]string, 0, len(groups))
	for group := range groups {
		groupNames = append(groupNames, group)
	}
	sort.Strings(groupNames)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	_, _ = fmt.Fprintln(w, "GROUP\tPROTOCOL TYPE")
	_, _ = fmt.Fprintln(w, "-----\t-------------")

	for _, group := range groupNames {
		_, _ = fmt.Fprintf(w, "%s\t%s\n", group, groups[group])
	}
	_ = w.Flush()

	return nil
}

func getTargetGroups(admin sarama.ClusterAdmin, groupFilter string) ([]string, error) {
	groups, err := admin.ListConsumerGroups()
	if err != nil {
		return nil, fmt.Errorf("failed to list consumer groups: %w", err)
	}

	targetGroups := make([]string, 0)
	if groupFilter != "" {
		if _, exists := groups[groupFilter]; !exists {
			return nil, fmt.Errorf("consumer group '%s' not found", groupFilter)
		}
		targetGroups = append(targetGroups, groupFilter)
	} else {
		for group := range groups {
			targetGroups = append(targetGroups, group)
		}
		sort.Strings(targetGroups)
	}
	return targetGroups, nil
}

func showConsumerGroupInfo(admin sarama.ClusterAdmin, client sarama.Client, groupFilter string) error {
	targetGroups, err := getTargetGroups(admin, groupFilter)
	if err != nil {
		return err
	}

	allInfo := make([]ConsumerGroupInfo, 0)
	totalLag := int64(0)

	for _, group := range targetGroups {
		groupInfo, err := getConsumerGroupInfo(admin, client, group)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to get info for group %s: %v\n", group, err)
			continue
		}

		for _, info := range groupInfo {
			allInfo = append(allInfo, info)
			totalLag += info.Lag
		}
	}

	if len(allInfo) == 0 {
		fmt.Println("No consumer group information available")
		return nil
	}

	// Print summary
	fmt.Printf("Total Consumer Groups: %d\n", len(targetGroups))
	fmt.Printf("Total Lag: %d messages\n", totalLag)
	fmt.Printf("Last Updated: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	// Print detailed information
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	_, _ = fmt.Fprintln(w, "GROUP\tTOPIC\tPARTITION\tCURRENT OFFSET\tLOG END OFFSET\tLAG\tCLIENT ID\tHOST")
	_, _ = fmt.Fprintln(w, "-----\t-----\t---------\t--------------\t--------------\t---\t---------\t----")

	for _, info := range allInfo {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d\t%d\t%s\t%s\n",
			info.Group,
			info.Topic,
			info.Partition,
			info.CurrentOffset,
			info.LogEndOffset,
			info.Lag,
			info.ClientID,
			info.Host,
		)
	}
	_ = w.Flush()

	return nil
}

func getConsumerGroupInfo(admin sarama.ClusterAdmin, client sarama.Client, group string) ([]ConsumerGroupInfo, error) {
	// Get group description
	descriptions, err := admin.DescribeConsumerGroups([]string{group})
	if err != nil {
		return nil, fmt.Errorf("failed to describe consumer group: %w", err)
	}

	if len(descriptions) == 0 {
		return nil, fmt.Errorf("no description found for group %s", group)
	}

	description := descriptions[0]

	// Get offset information
	offsetFetchRequest := &sarama.OffsetFetchRequest{
		Version:       1,
		ConsumerGroup: group,
	}

	coordinator, err := client.Coordinator(group)
	if err != nil {
		return nil, fmt.Errorf("failed to get coordinator: %w", err)
	}

	// Get list of topics
	topics, err := client.Topics()
	if err != nil {
		return nil, fmt.Errorf("failed to get topics: %w", err)
	}

	// Add all topic partitions to the request
	for _, topic := range topics {
		partitions, err := client.Partitions(topic)
		if err != nil {
			continue
		}
		for _, partition := range partitions {
			offsetFetchRequest.AddPartition(topic, partition)
		}
	}

	offsetFetchResponse, err := coordinator.FetchOffset(offsetFetchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch offsets: %w", err)
	}

	result := make([]ConsumerGroupInfo, 0)

	for topic, partitions := range offsetFetchResponse.Blocks {
		for partition, block := range partitions {
			if block.Offset < 0 {
				// No offset stored for this partition
				continue
			}

			// Get the log end offset
			logEndOffset, err := client.GetOffset(topic, partition, sarama.OffsetNewest)
			if err != nil {
				continue
			}

			lag := logEndOffset - block.Offset

			// Find member information
			clientID := ""
			host := ""
			memberID := ""
			groupInstanceID := ""
			for _, member := range description.Members {
				assignment, err := member.GetMemberAssignment()
				if err != nil || assignment == nil {
					continue
				}
				for assignedTopic, assignedPartitions := range assignment.Topics {
					if assignedTopic == topic {
						for _, assignedPartition := range assignedPartitions {
							if assignedPartition == partition {
								clientID = member.ClientId
								host = member.ClientHost
								memberID = member.MemberId
								if member.GroupInstanceId != nil {
									groupInstanceID = *member.GroupInstanceId
								}
								break
							}
						}
					}
				}
			}

			info := ConsumerGroupInfo{
				Group:           group,
				Topic:           topic,
				Partition:       partition,
				CurrentOffset:   block.Offset,
				LogEndOffset:    logEndOffset,
				Lag:             lag,
				ClientID:        clientID,
				Host:            host,
				MemberID:        memberID,
				GroupInstanceID: groupInstanceID,
			}

			result = append(result, info)
		}
	}

	// Sort by topic and partition
	sort.Slice(result, func(i, j int) bool {
		if result[i].Topic != result[j].Topic {
			return result[i].Topic < result[j].Topic
		}
		return result[i].Partition < result[j].Partition
	})

	return result, nil
}

func showMemberDetailInfo(admin sarama.ClusterAdmin, groupFilter string) error {
	targetGroups, err := getTargetGroups(admin, groupFilter)
	if err != nil {
		return err
	}

	for _, group := range targetGroups {
		if err := showGroupMemberDetails(admin, group); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to get details for group %s: %v\n\n", group, err)
			continue
		}
	}

	return nil
}

func showGroupMemberDetails(admin sarama.ClusterAdmin, group string) error {
	descriptions, err := admin.DescribeConsumerGroups([]string{group})
	if err != nil {
		return fmt.Errorf("failed to describe consumer group: %w", err)
	}

	if len(descriptions) == 0 {
		return fmt.Errorf("no description found for group %s", group)
	}

	description := descriptions[0]

	fmt.Printf("═══════════════════════════════════════════════════════════════════\n")
	fmt.Printf("Consumer Group: %s\n", group)
	fmt.Printf("State: %s\n", description.State)
	fmt.Printf("Protocol Type: %s\n", description.ProtocolType)
	fmt.Printf("Protocol: %s\n", description.Protocol)
	fmt.Printf("Members: %d\n", len(description.Members))
	fmt.Printf("═══════════════════════════════════════════════════════════════════\n\n")

	if len(description.Members) == 0 {
		fmt.Println("No active members")
		return nil
	}

	memberIdx := 1
	for _, member := range description.Members {
		fmt.Printf("─── Member %d ───\n", memberIdx)
		memberIdx++
		fmt.Printf("  Member ID:        %s\n", member.MemberId)
		fmt.Printf("  Client ID:        %s\n", member.ClientId)
		fmt.Printf("  Client Host:      %s\n", member.ClientHost)

		if member.GroupInstanceId != nil {
			fmt.Printf("  Group Instance:   %s\n", *member.GroupInstanceId)
		} else {
			fmt.Printf("  Group Instance:   (dynamic membership)\n")
		}

		// Get member assignment
		assignment, err := member.GetMemberAssignment()
		if err != nil || assignment == nil {
			if err != nil {
				fmt.Printf("  Assignment:       Error: %v\n", err)
			} else {
				fmt.Printf("  Assignment:       None\n")
			}
		} else {
			fmt.Printf("  Assigned Topics:  %d\n", len(assignment.Topics))
			totalPartitions := 0
			for _, partitions := range assignment.Topics {
				totalPartitions += len(partitions)
			}
			fmt.Printf("  Total Partitions: %d\n", totalPartitions)

			if len(assignment.Topics) > 0 {
				fmt.Printf("\n  Topic Assignments:\n")
				// Sort topics for consistent output
				topics := make([]string, 0, len(assignment.Topics))
				for topic := range assignment.Topics {
					topics = append(topics, topic)
				}
				sort.Strings(topics)

				for _, topic := range topics {
					partitions := assignment.Topics[topic]
					sort.Slice(partitions, func(i, j int) bool {
						return partitions[i] < partitions[j]
					})
					fmt.Printf("    • %s: %v\n", topic, partitions)
				}
			}
		}

		// Get member metadata
		metadata, err := member.GetMemberMetadata()
		if err != nil {
			fmt.Printf("  Metadata:         Error: %v\n", err)
		} else {
			if len(metadata.Topics) > 0 {
				fmt.Printf("\n  Subscribed Topics: %d\n", len(metadata.Topics))
				sort.Strings(metadata.Topics)
				for _, topic := range metadata.Topics {
					fmt.Printf("    • %s\n", topic)
				}
			}
			if len(metadata.UserData) > 0 {
				fmt.Printf("  User Data:        %d bytes\n", len(metadata.UserData))
			}
		}

		fmt.Println()
	}

	return nil
}

// Handler for consumer group
type consumerGroupHandler struct{}

func (h consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h consumerGroupHandler) ConsumeClaim(_ sarama.ConsumerGroupSession, _ sarama.ConsumerGroupClaim) error {
	return nil
}

var _ sarama.ConsumerGroupHandler = (*consumerGroupHandler)(nil)

// Web server
var (
	globalAdmin  sarama.ClusterAdmin
	globalClient sarama.Client
)

func startWebServer(admin sarama.ClusterAdmin, client sarama.Client) {
	globalAdmin = admin
	globalClient = client

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/group/", handleGroupPage)
	http.HandleFunc("/api/groups-summary", handleGroupsSummary)
	http.HandleFunc("/api/group-detail/", handleGroupDetailContent)

	addr := ":" + *webPort
	log.Printf("Starting web server on http://localhost%s\n", addr)
	log.Printf("Open your browser and navigate to http://localhost%s\n", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	// Redirect if path is not exactly "/"
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	html := `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Kafka Consumer Groups</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: #f5f5f5;
            padding: 20px;
        }
        .container { max-width: 1400px; margin: 0 auto; }
        h1 {
            color: #333;
            margin-bottom: 10px;
        }
        .subtitle {
            color: #666;
            font-size: 14px;
            margin-bottom: 20px;
        }
        .refresh-info {
            background: #e3f2fd;
            padding: 10px 15px;
            border-radius: 5px;
            margin-bottom: 20px;
            color: #1976d2;
            font-size: 14px;
        }
        .filter-container {
            background: white;
            padding: 15px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            margin-bottom: 20px;
            display: flex;
            gap: 15px;
            align-items: center;
        }
        .filter-input {
            flex: 1;
            max-width: 400px;
            padding: 10px 15px;
            border: 2px solid #e0e0e0;
            border-radius: 5px;
            font-size: 14px;
            transition: border-color 0.2s;
        }
        .filter-input:focus {
            outline: none;
            border-color: #4CAF50;
        }
        .filter-stats {
            color: #666;
            font-size: 13px;
        }
        .clear-filter {
            padding: 8px 15px;
            background: #f44336;
            color: white;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-size: 13px;
            transition: background 0.2s;
        }
        .clear-filter:hover {
            background: #d32f2f;
        }
        .groups-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
            gap: 20px;
        }
        .group-card {
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            padding: 20px;
            cursor: pointer;
            transition: all 0.2s;
            text-decoration: none;
            color: inherit;
            display: block;
        }
        .group-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
        }
        .group-card.filtered-out {
            display: none !important;
        }
        .group-name {
            font-size: 16px;
            font-weight: 600;
            color: #333;
            margin-bottom: 15px;
            word-break: break-word;
        }
        .group-stats {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 10px;
            font-size: 13px;
        }
        .stat-item {
            display: flex;
            flex-direction: column;
        }
        .stat-label {
            color: #666;
            font-size: 11px;
            text-transform: uppercase;
            margin-bottom: 4px;
        }
        .stat-value {
            font-weight: 600;
            font-size: 16px;
        }
        .lag-high { color: #d32f2f; }
        .lag-medium { color: #f57c00; }
        .lag-low { color: #388e3c; }
        .loading {
            text-align: center;
            padding: 40px;
            color: #999;
        }
    </style>
    <script>
        function filterGroups() {
            const input = document.getElementById('groupFilter');
            const filter = input.value.toLowerCase();
            const cards = document.querySelectorAll('.group-card');

            let visibleCount = 0;
            cards.forEach(card => {
                const groupName = card.querySelector('.group-name').textContent.toLowerCase();
                if (groupName.includes(filter)) {
                    card.classList.remove('filtered-out');
                    visibleCount++;
                } else {
                    card.classList.add('filtered-out');
                }
            });

            const stats = document.getElementById('filterStats');
            if (stats) {
                if (visibleCount === cards.length) {
                    stats.textContent = cards.length + ' groups';
                } else {
                    stats.textContent = visibleCount + ' / ' + cards.length + ' groups';
                }
            }
        }

        function clearFilter() {
            document.getElementById('groupFilter').value = '';
            filterGroups();
        }
    </script>
</head>
<body>
    <div class="container">
        <h1>Kafka Consumer Groups</h1>
        <div class="subtitle">Click a group to view detailed information</div>

        <div class="refresh-info">
            ⏱ Auto-refresh every ` + fmt.Sprintf("%d", *refreshInterval) + ` seconds | Broker: ` + *brokers + `
        </div>

        <div class="filter-container">
            <input type="text"
                   id="groupFilter"
                   class="filter-input"
                   placeholder="🔍 Filter by group name..."
                   oninput="filterGroups()"
                   autocomplete="off">
            <span class="filter-stats" id="filterStats">0 groups</span>
            <button class="clear-filter" onclick="clearFilter()">Clear</button>
        </div>

        <div class="groups-grid"
             hx-get="/api/groups-summary"
             hx-trigger="load, every ` + fmt.Sprintf("%ds", *refreshInterval) + `"
             hx-swap="innerHTML">
            <div class="loading">Loading consumer groups...</div>
        </div>
    </div>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprint(w, html)
}

func handleGroupsSummary(w http.ResponseWriter, _ *http.Request) {
	groups, err := globalAdmin.ListConsumerGroups()
	if err != nil {
		http.Error(w, "Failed to list groups", http.StatusInternalServerError)
		return
	}

	groupNames := make([]string, 0, len(groups))
	for group := range groups {
		groupNames = append(groupNames, group)
	}
	sort.Strings(groupNames)

	type GroupSummary struct {
		Name       string
		TotalLag   int64
		Partitions int
		Topics     int
		Members    int
	}

	summaries := make([]GroupSummary, 0)

	for _, group := range groupNames {
		groupInfo, err := getConsumerGroupInfo(globalAdmin, globalClient, group)
		if err != nil {
			continue
		}

		descriptions, _ := globalAdmin.DescribeConsumerGroups([]string{group})
		memberCount := 0
		if len(descriptions) > 0 {
			memberCount = len(descriptions[0].Members)
		}

		totalLag := int64(0)
		topicsSet := make(map[string]bool)
		for _, info := range groupInfo {
			totalLag += info.Lag
			topicsSet[info.Topic] = true
		}

		summaries = append(summaries, GroupSummary{
			Name:       group,
			TotalLag:   totalLag,
			Partitions: len(groupInfo),
			Topics:     len(topicsSet),
			Members:    memberCount,
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	for _, summary := range summaries {
		lagClass := "lag-low"
		if summary.TotalLag > 100000 {
			lagClass = "lag-high"
		} else if summary.TotalLag > 10000 {
			lagClass = "lag-medium"
		}

		_, _ = fmt.Fprintf(w, `<a href="/group/%s" class="group-card">
			<div class="group-name">%s</div>
			<div class="group-stats">
				<div class="stat-item">
					<div class="stat-label">Total Lag</div>
					<div class="stat-value %s">%s</div>
				</div>
				<div class="stat-item">
					<div class="stat-label">Partitions</div>
					<div class="stat-value">%d</div>
				</div>
				<div class="stat-item">
					<div class="stat-label">Topics</div>
					<div class="stat-value">%d</div>
				</div>
				<div class="stat-item">
					<div class="stat-label">Members</div>
					<div class="stat-value">%d</div>
				</div>
			</div>
		</a>`,
			summary.Name,
			summary.Name,
			lagClass, formatNumber(summary.TotalLag),
			summary.Partitions,
			summary.Topics,
			summary.Members)
	}

	_, _ = fmt.Fprintf(w, `<script>
		setTimeout(() => {
			const totalGroups = %d;
			const stats = document.getElementById('filterStats');
			if (stats) stats.textContent = totalGroups + ' groups';
		}, 50);
	</script>`, len(summaries))
}

func formatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	} else if n < 1000000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}

	return fmt.Sprintf("%.1fM", float64(n)/1000000)
}

func handleGroupPage(w http.ResponseWriter, r *http.Request) {
	groupName := strings.TrimPrefix(r.URL.Path, "/group/")
	if groupName == "" {
		http.Error(w, "Group name required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + groupName + ` - Kafka Consumer Group</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: #f5f5f5;
            padding: 20px;
        }
        .container { max-width: 1600px; margin: 0 auto; }
        .header {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            margin-bottom: 20px;
        }
        .back-button {
            display: inline-block;
            padding: 8px 16px;
            background: #666;
            color: white;
            text-decoration: none;
            border-radius: 5px;
            margin-bottom: 15px;
            transition: background 0.2s;
        }
        .back-button:hover {
            background: #555;
        }
        h1 {
            color: #333;
            margin-bottom: 15px;
        }
        .refresh-info {
            background: #e3f2fd;
            padding: 10px 15px;
            border-radius: 5px;
            margin-top: 15px;
            color: #1976d2;
            font-size: 14px;
        }
        .group-meta {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            padding: 15px;
            background: #f9f9f9;
            border-radius: 5px;
        }
        .meta-item {
            display: flex;
            flex-direction: column;
        }
        .meta-label {
            color: #666;
            font-size: 12px;
            text-transform: uppercase;
            margin-bottom: 5px;
        }
        .meta-value {
            font-weight: 600;
            font-size: 16px;
        }
        .lag-high { color: #d32f2f; }
        .lag-medium { color: #f57c00; }
        .lag-low { color: #388e3c; }
        .section {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            margin-bottom: 20px;
        }
        h2 {
            color: #333;
            margin-bottom: 15px;
            padding-bottom: 10px;
            border-bottom: 2px solid #4CAF50;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            font-size: 13px;
        }
        thead {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
        }
        th {
            padding: 12px 8px;
            text-align: left;
            font-weight: 600;
        }
        tbody tr {
            border-bottom: 1px solid #eee;
        }
        tbody tr:hover {
            background: #f5f9ff;
        }
        td {
            padding: 10px 8px;
        }
        .member-card {
            background: #f9f9f9;
            border-left: 4px solid #4CAF50;
            padding: 15px;
            margin: 10px 0;
            border-radius: 5px;
        }
        .member-header {
            font-weight: bold;
            color: #333;
            margin-bottom: 10px;
        }
        .member-info {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 10px;
            font-size: 13px;
            margin-bottom: 10px;
        }
        .info-label {
            color: #666;
            font-weight: 500;
        }
        .topic-assignments {
            margin-top: 10px;
            padding: 10px;
            background: white;
            border-radius: 5px;
        }
        .topic-item {
            padding: 5px;
            margin: 3px 0;
            font-size: 12px;
            background: #e8f5e9;
            border-radius: 3px;
        }
        .loading {
            text-align: center;
            padding: 40px;
            color: #999;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <a href="/" class="back-button">← Back to Groups</a>
            <h1>` + groupName + `</h1>
            <div class="refresh-info">
                ⏱ Auto-refresh every ` + fmt.Sprintf("%d", *refreshInterval) + ` seconds
            </div>
        </div>

        <div id="group-content"
             hx-get="/api/group-detail/` + groupName + `"
             hx-trigger="load, every ` + fmt.Sprintf("%ds", *refreshInterval) + `"
             hx-swap="innerHTML">
            <div class="loading">Loading group details...</div>
        </div>
    </div>
</body>
</html>`

	_, _ = fmt.Fprint(w, html)
}

func handleGroupDetailContent(w http.ResponseWriter, r *http.Request) {
	groupName := strings.TrimPrefix(r.URL.Path, "/api/group-detail/")
	if groupName == "" {
		http.Error(w, "Group name required", http.StatusBadRequest)
		return
	}

	// Get group info
	groupInfo, err := getConsumerGroupInfo(globalAdmin, globalClient, groupName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get group info: %v", err), http.StatusInternalServerError)
		return
	}

	descriptions, err := globalAdmin.DescribeConsumerGroups([]string{groupName})
	if err != nil || len(descriptions) == 0 {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	description := descriptions[0]

	// Calculate total lag
	totalLag := int64(0)
	for _, info := range groupInfo {
		totalLag += info.Lag
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Group metadata
	html := `<div class="group-meta">
        <div class="meta-item">
            <div class="meta-label">State</div>
            <div class="meta-value">` + description.State + `</div>
        </div>
        <div class="meta-item">
            <div class="meta-label">Protocol</div>
            <div class="meta-value">` + description.Protocol + `</div>
        </div>
        <div class="meta-item">
            <div class="meta-label">Total Lag</div>
            <div class="meta-value ` + getLagClass(totalLag) + `">` + fmt.Sprintf("%d", totalLag) + `</div>
        </div>
        <div class="meta-item">
            <div class="meta-label">Partitions</div>
            <div class="meta-value">` + fmt.Sprintf("%d", len(groupInfo)) + `</div>
        </div>
        <div class="meta-item">
            <div class="meta-label">Members</div>
            <div class="meta-value">` + fmt.Sprintf("%d", len(description.Members)) + `</div>
        </div>
    </div>

    <div class="section">
        <h2>Partitions</h2>
        <table>
            <thead>
                <tr>
                    <th>TOPIC</th>
                    <th>PARTITION</th>
                    <th>CURRENT OFFSET</th>
                    <th>LOG END OFFSET</th>
                    <th>LAG</th>
                    <th>CLIENT ID</th>
                    <th>HOST</th>
                </tr>
            </thead>
            <tbody>`

	for _, info := range groupInfo {
		lagClass := getLagClass(info.Lag)
		html += fmt.Sprintf(`
                <tr>
                    <td>%s</td>
                    <td>%d</td>
                    <td>%d</td>
                    <td>%d</td>
                    <td class="%s">%d</td>
                    <td>%s</td>
                    <td>%s</td>
                </tr>`,
			info.Topic,
			info.Partition,
			info.CurrentOffset,
			info.LogEndOffset,
			lagClass, info.Lag,
			info.ClientID,
			info.Host)
	}

	html += `
            </tbody>
        </table>
    </div>

    <div class="section">
        <h2>Members</h2>`

	if len(description.Members) == 0 {
		html += `<div style="text-align: center; padding: 20px; color: #999;">No active members</div>`
	} else {
		memberIdx := 1
		for _, member := range description.Members {
			assignment, _ := member.GetMemberAssignment()

			totalPartitions := 0
			if assignment != nil {
				for _, partitions := range assignment.Topics {
					totalPartitions += len(partitions)
				}
			}

			groupInstance := "(dynamic)"
			if member.GroupInstanceId != nil {
				groupInstance = *member.GroupInstanceId
			}

			html += fmt.Sprintf(`
        <div class="member-card">
            <div class="member-header">Member %d - %s</div>
            <div class="member-info">
                <div><span class="info-label">Member ID:</span> %s</div>
                <div><span class="info-label">Client ID:</span> %s</div>
                <div><span class="info-label">Host:</span> %s</div>
                <div><span class="info-label">Instance:</span> %s</div>
                <div><span class="info-label">Partitions:</span> %d</div>
            </div>`,
				memberIdx, member.ClientId, member.MemberId, member.ClientId, member.ClientHost, groupInstance, totalPartitions)

			if assignment != nil && len(assignment.Topics) > 0 {
				html += `<div class="topic-assignments"><strong>Assigned Topics:</strong><br>`
				topics := make([]string, 0, len(assignment.Topics))
				for topic := range assignment.Topics {
					topics = append(topics, topic)
				}
				sort.Strings(topics)

				for _, topic := range topics {
					partitions := assignment.Topics[topic]
					sort.Slice(partitions, func(i, j int) bool {
						return partitions[i] < partitions[j]
					})
					partitionsJSON, _ := json.Marshal(partitions)
					html += fmt.Sprintf(`<div class="topic-item">%s: %s</div>`, topic, string(partitionsJSON))
				}
				html += `</div>`
			}

			html += `</div>`
			memberIdx++
		}
	}

	html += `</div>`

	_, _ = fmt.Fprint(w, html)
}

func getLagClass(lag int64) string {
	if lag > 100000 {
		return "lag-high"
	} else if lag > 10000 {
		return "lag-medium"
	}
	return "lag-low"
}
