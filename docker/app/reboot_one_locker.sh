#!/bin/bash

if [ "$#" -ne 3 ]; then
  echo "Usage: $0 <LOCKER_Id> <PASSWORD> <IP>"
  exit 1
fi

LOCKER_Id="$1"
PASSWORD="$2"
IP="$3"

sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -tt user@"$IP" "./reboot_one_locker.sh $LOCKER_Id"

