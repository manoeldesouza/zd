# zd
zd is a data protection software for ZFS. zd handles zpool scrubs, dataset and volume snapshots and replication (send | receive) according to the schedule and definitions from YAML configuration file.


## Features
- single yaml configuration file (default zd.yaml)
- zpool scrubs (daily, weekly, monthly)
- dataset and volume snapshots (hourly, daily, weekly, monthly)
- same dataset or volume may have different snapshot frequencies
- local snapshot pruning according to different retention policies
- recursive snapshots
- local and remote replication
- recursive replication


## Installation
Decompress the contents for your system (freebsd.zip or linux.zip) of build into `/usr/local/bin` or the appropriate bin diretory present in your `PATH` variable.


## Usage
`zd [-c config_file]` as root


## Configuration

```
---
tasks:
  - task: scrub
    target: Pool1/Dataset1
    frequency: weekly
    when: Sat 05:00       # Monthly schedule is <DayOfMonth> HH:MM
                          # Weekly schedule is <DayOfWeek> HH:MM
                          # Daily schedule is HH:MM

  - task: snapshot
    target: Pool2/Dataset2
    frequency: hourly
    when: 00              # Hourly schedule is MM
    recursive: true       # Snapshot child datasets
    retention: 4          # Number of snapshots of this frequency for this 
                          # dataset to preserve. Older snapshots are pruned

  - task: replication
    target: Pool3/Dataset3
    frequency: daily
    when: 00:00
    recursive: false
    raw: true             # Preserve the data as stored. When true encrypted filesystems 
                          # are not decrypted before replicated. Local key will be
                          # required on the destination 
    to: 
      # Local dataset for the replication
      - Pool4/Backup

      # remote dataset for the replication. format is <dataset>@<hostname|ip address>
      - Pool5/Backup/Media@remote_host
...
```


## Compilation

For FreeBSD
```
$ make
$ cp build/zd /usr/local/bin
$ chown root:wheel /usr/local/bin/zd
```

For Linux
```
$ make
$ cp build/zd /usr/bin
$ chown root:root /usr/bin/zd
```


## Copyright

zd is copyright to Manoel de Souza 2023


