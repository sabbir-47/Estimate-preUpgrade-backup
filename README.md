# Estimate-preUpgrade-backup  

This utility prints the estimate of pre-upgrade backup disk size and actual disk size used after backup has taken to compare the correctness of the estimation.

# Usage

* Executing the binary without passing any parameter will print `Pre-backup estimated disk size`
* Executing the binary with actual parameter will print `Pre-backup estimated disk size` and `Post-backup actual disk used`

An example is provided below:

```

[root@lab-test-spoke2-node-0 core]# ./main actual

******************************
Pre-backup estimated disk size 
******************************
--------------------------------------------------------------------------------
Resource  | Directory                    | Size      | Percentage|
cluster   | /var/lib/etcd/member/snap/db | 95.78 MiB | 1.52%     |
usrLocal  | /usr/local                   | 15.00 B   | 0.00%     |
kubelet   | /var/lib/kubelet             | 5.96 GiB  | 96.97%    |
etc       | /etc                         | 71.35 MiB | 1.13%     |
--------------------------------------------------------------------------------
                             TOTAL =   6.15 GiB


****************************
Post-backup actual disk used 
****************************
--------------------------------------------------------------------------------
Resource            | Directory                         | Size      | Percentage|
upgrade-recovery.sh | /var/recovery/upgrade-recovery.sh | 14.56 KiB | 0.00%     |
cluster             | /var/recovery/cluster             | 90.04 MiB | 1.40%     |
etc.exclude.list    | /var/recovery/etc.exclude.list    | 160.00 B  | 0.00%     |
etc                 | /var/recovery/etc                 | 71.17 MiB | 1.11%     |
usrlocal            | /var/recovery/usrlocal            | 36.15 KiB | 0.00%     |
kubelet             | /var/recovery/kubelet             | 6.12 GiB  | 97.49%    |
--------------------------------------------------------------------------------
                             TOTAL =   6.28 GiB

```
