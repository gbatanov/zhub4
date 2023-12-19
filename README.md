## Zhub4

Golang version of my project home automation system on C++ (zhub2)
<<<<<<< HEAD
=======

>>>>>>> master

Starting from version 0.5.34, the project is developed only for notebooks and desktop computers.
Внимание! На сегодня версия не полностью рабочая! использовать только в ознакомительных целях!

##### Directories:
- httpServer - HTTP server, web api to zigbee part
- serial3 - serial port package
- telega32 - telegram bot package
- zigbee - package for working with zigbee
  - zdo - main zigbee package
    - zcl - basic concepts and zigbee constants
  - clusters - a package of handlers in clusters

#### Settings
File with config need place into /usr/local/etc/zhub4 (content example in  config_example.txt)

#### Build and start (Linux, Mac)
```
make
sudo make install
sudo zhub4
```

#### Console commands
- q Quit
- j Permit join

<p>Copyright(c) GSB, Georgii Batanov, 2022-2023<br>
gbatanov@yandex.ru</p>
