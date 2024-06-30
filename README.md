## Zhub4

Golang version of my project home automation system on C++ (zhub2)

Starting from version 0.5.34, the project is developed only for notebooks and desktop computers.
Внимание! Использовать только в ознакомительных целях!

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

#### Для установки порта на CH341
```
sudo systemctl stop brltty-udev.service
sudo systemctl mask brltty-udev.service
sudo systemctl stop brltty.service
sudo systemctl disable brltty.service
```

#### Build and start (Linux, Mac)
```
make
sudo make install
sudo zhub4
```

#### Console commands
- q Quit
- j Permit join

#### History
Since version 12, PostgreSQL is used instead of a configuration file.

<p>Copyright(c) GSB, Georgii Batanov, 2022-2024<br>
gbatanov@yandex.ru</p>
