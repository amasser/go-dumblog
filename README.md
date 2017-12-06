Dumblog
=======

Dumblog will write a message to an AWS Cloudwatch Logs group and stream given
as arguments.  It follows the standard credential fallback path as the AWS SDK.

This code was purpose built to solve a pressing issue in a legacy system. It is
not suitable for writing logs where you can't afford to lose some of the messages
or you if you need to write a lot of messages frequently

Usage
=====

```
dumblog -group /aws/testgroup -stream some-unique-name 'My message'
```
