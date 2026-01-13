# Skaldenmet

Simple, single-node program to track resources on the server. It provides simple monitoring tools to
launch and monitor GPU and CPU usage for the single-node jobs. It tracks the resources with the PGID
allowing to monitor MPI type of jobs. Currently it uses NVIDIA NVML bindings to monitor the GPU usage.
