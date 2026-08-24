build:
    docker build  apps/smart_home -t smarthome:latest

run:
    docker run -d  --name smarthome smarthome:latest

stop:
    docker stop smarthome

logs:
    docker logs smarthome
# отрендерить все .puml в PNG рядом с исходниками (заменяет старые)
render:
    find schemas -name '*.puml' -print0 | xargs -0 plantuml -tpng -charset UTF-8
