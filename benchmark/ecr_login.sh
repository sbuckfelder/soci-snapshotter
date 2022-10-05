ada credentials update --account=596114853808 --provider=isengard --role=Admin --once
aws ecr get-login-password --endpoint https://api.starport.us-west-2.amazonaws.com --region us-west-2 > aws_secret
cat aws_secret | sudo docker login -u AWS --password-stdin 596114853808.dkr.starport.us-west-2.amazonaws.com
