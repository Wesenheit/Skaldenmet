# Skaldenmet

A simple, single-node program to track resources on servers in a SLURM-like manner.
 It is designed as a simple, easy-to-configure tool that can launch and monitor resource usage.
 The main goal of this application is to facilitate HPC/ML resource tracking on single-node private workstations.

It is designed to have similar usage to the SLURM scheduler while being much easier to configure and maintain for solo users.
 It tracks resources of jobs together with all child processes, allowing it to monitor MPI or any other type of multi-process jobs.
 It supports CPU and GPU monitoring of NVIDIA GPUs with the NVML library, making it ideal for tracking the resources of ML jobs.

## Installation

To install the application, you can either build it from source or install it with:
```bash
go install github.com/Wesenheit/Skaldenmet/cmd/met@latest
```

## Usage

There are three subapplications in the current release: daemon, runner, and display tool.

### Daemon

To collect system-level performance metrics, you need to run the daemon in the background. To do so, type:
```bash
met daemon --config config_file.yaml
```
Examples of configuration files can be found in the `examples` directory.
 The detailed meaning of the config can be found at the end of this README.

### Runner

Jobs are submitted with the runner module. To launch and monitor consumed resources, type:
```bash
$ met run --name some_job -- ./command/to/execute
2026/01/14 16:40:44 Started command some_job with PPID 3113
```
This will launch a job and redirect standard output to `some_job.out` and standard error to `some_job.err`.
 All environmental variables are inherited by the process, allowing seamless integration with existing workflows.

### Display Tool

To display the results of jobs, there is a CLI tool that shows all running and finished jobs with associated performance measures.
 If CPU monitoring is enabled, you can retrieve the information by running:

```bash
$ met list cpu
┌──────┬──────────────┬────────────────┬────────────────┬────────┬──────────┐
│ PPID │ PROCESS NAME │ CPU  % ( AVG ) │ MEM  % ( AVG ) │ STATUS │ DURATION │
├──────┼──────────────┼────────────────┼────────────────┼────────┼──────────┤
│ 3113 │ some_job     │ 376.41%        │ 0.01%          │ Active │ 1m8s     │
└──────┴──────────────┴────────────────┴────────────────┴────────┴──────────┘
```

If the GPU collector is enabled, you can see all GPU stats by running:
```bash
$ met list gpu
┌──────┬──────────────┬───────────────────┬──────────────┬─────────────┬──────────┬────────┬──────────┐
│ PPID │ PROCESS NAME │ GPU UTIL  ( AVG ) │ MEM  ( AVG ) │ TOTAL POWER │ MAX TEMP │ STATUS │ DURATION │
├──────┼──────────────┼───────────────────┼──────────────┼─────────────┼──────────┼────────┼──────────┤
│ 4906 │ some_job     │ 12.31%            │ 12.37 GB     │ 0.69 Wh     │ 53.00 C  │ Active │ 48s      │
└──────┴──────────────┴───────────────────┴──────────────┴─────────────┴──────────┴────────┴──────────┘
```

## Documentation & Design

`Skaldenmet` was designed to be as simple and easy to configure as possible.
 As such, the entire configuration is done with a single YAML configuration file with sections dedicated to various components.

### Storage

Currently, only one simple memory storage is supported. Future plans include SQLite-based storage.
Hence, there is a configuration section dedicated to storage options:

```yaml
storage:
  name: "memory"
  size: 100
  interval: "4s"
```

This configuration will create a memory-based storage with 100 slots that will be aggregated every 4 seconds.
For the current configuration, size is the most important parameter, as too many records can consume system memory (albeit the memory footprint is small). In the current version, only a single aggregated record is maintained; individual measurements are discarded.

### Collectors

Collectors are submodules responsible for resource collection. As such, they are configured independently.
Currently, there are two collectors: CPU and GPU (NVIDIA). Both are configured in the same way:

```yaml
cpuCollector:
  interval: "1s"
  size: 10
nvidiaCollector:
  interval: "0.5s"
  size: 10
```

The `interval` parameter controls how frequently resources are queried.
The `size` parameter controls the internal memory storage for the module; after exceeding local storage,
 measurements are moved to the main storage for aggregation.

### Internal Process Mapping

Collectors do not query resources randomly; they only target processes that were spawned from the main process.
The main process is spawned with a certain PGID, which is also the PPID for the group leader.
To reliably look for all processes with a given PGID on various platforms (macOS and Linux), an internal mapping between PGIDs and PPIDs is maintained.

This internal mapping needs to be refreshed to remove completed processes and add new processes that were spawned.
To find new processes belonging to a given PGID, there is a need to iterate over all existing processes.
Hence, it can be time- and resource-consuming and is done only periodically. The frequency of this lookup can be controlled with the parameter:

```yaml
state:
  interval: "2s"
```
## Roadmap

- [ ] SQLite-based persistent storage
- [ ] Support for AMD GPUs
- [ ] TUI
