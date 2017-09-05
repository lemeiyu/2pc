2pc
======
This tool allows you to run script on any specified node in a consul cluster, while ensuring the uniformity of these nodes.

You need to provide:
------
**_task name_**, the name of your task, if you don't want to name your task, it will get a default name.

**_action script_**, which is the script you want to run on each node, and you must provide arguments the action script needs (unless script don't need any argument); 

**_rollback script_**, the script can undo the work done by the action script, arguments are required as well.

**_timeout_**, the unit of timeout is second.

Notes
------
Because of the limit of consul, the script file cannot lager than **512 KB**.

Updates
------
Now can run script other than bash or even a program file.

New task cannot start until former task finished.

You don't need to provide the task directory which saves scripts, a temp dir will be created to store scripts. That means you must use absolute path in your script.

How to use this tool
------
1.You will need to build and put the scriptLoader, which is the executor of scripts, to each node in your consul cluster. Make sure scriptLoader is available on the PATH

2.Import two-pc package into your file. Create a TaskInfo struct **t**, initial it by **t**._New()_ function with all required information about your task.

3.Use **t**._RunAction()_ to start your task.