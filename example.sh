#/bin/bash

# AWS_ACCESS_KEY_ID example (fake)
export AWS_ACCESS_KEY_ID=KkRbWpoyqLHo69dvoskn 

# AWS_SECRET_ACCESS_KEY example (fake)
export AWS_SECRET_ACCESS_KEY=VXFvVu0ckOYP22WYxAz9TyoNYC8Rx1WVLVFK3Fyc

# AWS_DEFAULT_REGION default region
export AWS_DEFAULT_REGION=us-west-2

go build

./route53ddns -name home.example.com -ttl 60 -hostedZoneID Z2DKEZEXAMPLE1


# crontab
This can be run other ways but crontab was the target.