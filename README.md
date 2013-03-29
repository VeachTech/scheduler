scheduler
=========
scheduler is meant to be a lightweight and flexible job scheduler

A Job is an interface that is made up of 3 parts:

* NextRunTime() Which takes a time.Time as input and returns the next time 
the Job should be run

* Run() takes no parameters and is the function that gets called when 
it is time to run the job

* GetID() returns an uint that is used to identify the job for updating or
removing the job from the scheduler

What's Included
---------------
* The scheduler itself

* A simple ID type that can be dropped into a struct to Satisfy 
the GetID() requirement for a Job

* A CronTime type that specifies NextRunTime() using the standard Cron Syntax
(except for starting with seconds instead of minutes)

To get started you only need to define a function (Run()) and drop it into a struct along with ID and CronTime (and initialize them) and you are ready to add your Job to the scheduler.

Of course you are free to implement your own GetID() or NextRunTime()