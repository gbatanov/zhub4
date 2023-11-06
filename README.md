## Zhub4

Golang version of my project home automation system on C++

Starting from version 0.5.34, the project is developed only for Linux.

##### Directories:
- http_server - HTTP server, web api to zigbee part
- serial3 - serial port package
- telega32 - telegram bot package
- zigbee - package for working with zigbee
  - zdo - main zigbee package
    - zcl - basic concepts and zigbee constants
  - clusters - a package of handlers in clusters

#### Settings
File with config need place into /usr/local/etc/zhub4 (content example in  config_example)

#### Build an start
```
make
make install
zhub4
```

#### Console commands
- q Quit
- j Permit join

<p>George Batanov, 2022-2023<br>
gbatanov@yandex.ru</p>
