The goal of this is to be a simple utility that automatically restarts a process when it or one of its children is killed.

I mainly wanted to make this because long-running processes on my university's lab machines get automatically killed. This utility allows processes such as servers to be automatically restarted because the main process enters a sleep state while waiting for signals from the child processes and doesn't use CPU time. Thus, the child processes get targeted and necr wakes up when they get killed.

