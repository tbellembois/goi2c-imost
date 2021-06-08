#!/usr/bin/env bash
facetteServerAddress=""
collectdCSVDataDir=""
mailServerAddress=""
mailServerPort=""
mailServerFrom=""
mailServerTo=""
debug=""

probes=$NOTIFIER_PROBES

if [ ! -z "$NOTIFIER_FACETTESERVERADDRESS" ]
then
      facetteServerAddress="-facetteServerAddress $NOTIFIER_FACETTESERVERADDRESS"
fi
if [ ! -z "$NOTIFIER_COLLECTDCSVDATADIR" ]
then
      collectdCSVDataDir="-collectdCSVDataDir $NOTIFIER_COLLECTDCSVDATADIR"
fi
if [ ! -z "$NOTIFIER_MAILSERVERADDRESS" ]
then
      mailServerAddress="-mailServerAddress $NOTIFIER_MAILSERVERADDRESS"
fi
if [ ! -z "$NOTIFIER_MAILSERVERPORT" ]
then
      mailServerPort="-mailServerPort $NOTIFIER_MAILSERVERPORT"
fi
if [ ! -z "$NOTIFIER_MAILSERVERFROM" ]
then
      mailServerFrom="-mailServerFrom $NOTIFIER_MAILSERVERFROM"
fi
if [ ! -z "$NOTIFIER_MAILSERVERTO" ]
then
      mailServerTo="-mailServerTo $NOTIFIER_MAILSERVERTO"
fi
if [ ! -z "$NOTIFIER_DEBUG" ]
then
      debug="-debug"
fi

/var/www-data/notifier \
$facetteServerAddress \
$collectdCSVDataDir \
$mailServerAddress \
$mailServerPort \
$mailServerFrom \
$mailServerTo \
$debug \
$probes

echo "Sleeping..."
# Spin until we receive a SIGTERM (e.g. from `docker stop`)
trap 'exit 143' SIGTERM # exit = 128 + 15 (SIGTERM)
tail -f /dev/null & wait ${!}