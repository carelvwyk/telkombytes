Telkom Bytes
===========

Telkom does not provide a nice interface to track Telkom Mobile internet usage over time. Luckily current usage information (bytes remaining on data cap) is available from an unauthenticated web portal when visited from a Telkom Mobile internet connection.

"Telkom Bytes" is intended to run on a device on my local network like a Raspberry Pi. It tracks my Telkom Mobile bundle usage and pushes the usage data to [AWS CloudWatch](https://aws.amazon.com/cloudwatch/). CloudWatch can graph my internet usage over time and alert me via SMS when my remaining balance drop below a threshold.

This project is Go based because I love Go.

TODO:
- Hook up to CloudWatch
- Set up on a Raspberry Pi