build:
    docker build  apps/smart_home -t smarthome:latest

run:
    docker run -d  --name smarthome smarthome:latest

stop:
    docker stop smarthome

logs:
    docker logs smarthome