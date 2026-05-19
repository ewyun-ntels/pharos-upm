export const sampleData = {
  timestamp: "2025-03-25T12:00:00Z",
  cpu: {
    usage: {
      total: 62.5,
      user: 40.3,
      system: 20.2,
      idle: 37.5,
      iowait: 2.0,
    },
    cores: [
      { core_id: 0, usage: 58.3 },
      { core_id: 1, usage: 65.1 },
      { core_id: 2, usage: 63.2 },
      { core_id: 3, usage: 61.5 },
    ],
    load_average: {
      "1m": 2.35,
      "5m": 1.89,
      "15m": 1.72,
    },
    temperature: {
      current: 65.2,
      critical: 85.0,
    },
    processes: [
      { pid: 1234, name: "nginx", cpu_usage: 15.2 },
      { pid: 5678, name: "java", cpu_usage: 25.8 },
      { pid: 9101, name: "postgres", cpu_usage: 10.5 },
    ],
    scheduling: {
      context_switches: 30456,
      interrupts: 14578,
    },
    clock_speed: {
      current: 3.2,
      max: 4.0,
    },
    threads: {
      total_threads: 256,
      active_cores: 4,
    },
    wait_time: {
      iowait: 2.0,
    },
  },
} as const;
