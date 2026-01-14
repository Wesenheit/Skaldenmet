# Skaldenmet

Simple, single-node program to track resources on the server in a SLURM-like manner. It is designed as a simple, easy to configure
tool that can launch and monitor the usage of resources. The main goal of this application is to facilitate HPC/ML
resource tracking on single-node private workstations. 
It is designed to have similar usage to SLURM scheduler while beeing much more easy to configure and maintain for solo
users.
It allows to track resources of the jobs together with all children processes.
As such, it allows to track the MPI or any other type of multi-process jobs. It supports CPU and GPU monitoring of NVIDIA GPU-s
with the NVML library. Hence, it is ideal to track the resources of ML jobs.

# Instalation
To install the application one can simply build it from source or installing with `go install https://github.com/Wesenheit/Skaldenmet/cmd/met/main.go`.

## Usage
There are three subapplications in the current realese: daemon, runner, display tool.

### Daemon
In order to collect the system-level performance metricks one needs to run the daemon in the background. To do so, type `met daemon --config conig_file.yaml`.
Examples of configuration files are can be found in `examples` directory. The detail meaning of the config can be found at
the end of this README.

### Runner
Jobs are submitted with the runner module. To launch and monitor the consumed resource type
```
$ met run --name some_job -- ./command/to/exectute
2026/01/14 16:40:44 Started command some_job with PPID 3113
  
```
This will launch a job and redirect standard output to `some_job.out` and standard error to `some_job.err`.
All enviromental variables are inherited by the process allowing seamless integration with existing workflows.

### Display tool
To display the results of the jobs there is a CLI tool that shows all runing and finished jobs with
associated performance measures. If the CPU monitoring is enabled, one can recover the information by running

```
$ met run list cpu
┌──────┬──────────────┬────────────────┬────────────────┬────────┬──────────┐
│ PPID │ PROCESS NAME │ CPU  % ( AVG ) │ MEM  % ( AVG ) │ STATUS │ DURATION │
├──────┼──────────────┼────────────────┼────────────────┼────────┼──────────┤
│ 3113 │ some_job     │ 376.41%        │ 0.01%          │ Active │ 1m8s     │
└──────┴──────────────┴────────────────┴────────────────┴────────┴──────────┘
  
```
If the GPU collector is enabled one can see all GPU stats by running
```
 $ met list gpu
┌──────┬──────────────┬───────────────────┬──────────────┬─────────────┬──────────┬────────┬──────────┐
│ PPID │ PROCESS NAME │ GPU UTIL  ( AVG ) │ MEM  ( AVG ) │ TOTAL POWER │ MAX TEMP │ STATUS │ DURATION │
├──────┼──────────────┼───────────────────┼──────────────┼─────────────┼──────────┼────────┼──────────┤
│ 4906 │ some_job     │ 12.31%            │ 12.37 GB     │ 0.69 Wh     │ 53.00 C  │ Active │ 48s      │
└──────┴──────────────┴───────────────────┴──────────────┴─────────────┴──────────┴────────┴──────────
````

# Documentation & Design
`Skaldenmet` was designed to be a simple and as easy configure as possible. As such, total configuration is done
with a single yaml configuration file with sections dedicated to various components.


## Storage
Currently ony one simple memory storage is supported. Future plans include SQLite based storage. Hence,
there is a configuration section dedicated to the storage options.
```
storage:
  name: "memory"
  size: 100
  interval: "4s"
```
Following config will create a memory based storage with 100 slots that will be aggregated every 4 seconds. For the current
configuration size is the most important one as to many records can eat into the ovarall system memory (albetit, the memory footprint is small).
In the current version only a single agregated record is maintained, single measurements are discarded.

## Collectors
Collectors are submodules that are responsible for the resource collection. As such, they are configured
independetly. Currently there are two collectors: CPU and GPU (NVIDIA). Both are configured in the same way.
```
cpuCollector:
  interval: "1s"
  size: 10
nvidiaCollector:
  interval: "0.5s"
  size: 10
```
Interval controls how frequently resources are queried. Size parameter controls the internal memory storage for the
module, after exceding local storage measurements are moved to the storage for aggregation.

## Internal process mapping
Collectors do not query resources randomly, they only target processes that were spawned from the main process.
Main process is spawned with certain PGID which is also a PPID for a group leader. In order to reliably look
for all processes with given PGID on various platforms (MacOS and Linux), an internal mapping between PGIDs and PPIDS is maintained.
This internal mapping needs to be refreshed to remove completed processes and
add new processes that were spawned. To find new processes that belong to the given PGID there is a need to iterate over all existing processes.
 Hence, it can be time and resource consuming and is done only every once in a while.
The frequency of this lookup can be controled with the paramter:
```
state:
  interval: "2s"
```



