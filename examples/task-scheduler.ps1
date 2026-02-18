# Windows Task Scheduler example
$Action = New-ScheduledTaskAction -Execute "moltbb" -Argument "run"
$Trigger = New-ScheduledTaskTrigger -Daily -At 9:00PM
Register-ScheduledTask -TaskName "MoltBBDiary" -Action $Action -Trigger $Trigger
