#!/bin/bash

# This script adds SAP-internal information to the documentation of piper-lib-os steps
# For a piper-lib-os step "toolExecute" that is documented in "toolExecute.md",
# SAP-internal information needs to be maintained in "docs/steps/__toolExecute_sap.md".
# This file will be appended to the end of "toolExecute.md".

if ls docs/steps/__*_sap.md 1> /dev/null 2>&1; then
  echo "Appending SAP-internal information to piper-lib-os steps from the following files:"

  for filepathSAPSpecifics in docs/steps/__*_sap.md; do
    filepathPiperOSStep=$(echo "$filepathSAPSpecifics" | sed -n 's/\/__\(.*\)_sap\.md/\/\1.md/p')

    # add sap-spcific marker
    sed -i '/## ${docGenParameters}/i ## ${sapSpecifics}\n' $filepathPiperOSStep && \
    # insert sap-specific content
    sed -i "/## \${sapSpecifics}/r $filepathSAPSpecifics" $filepathPiperOSStep && \
    # remove sap-spcific marker
    sed -i '/## ${sapSpecifics}/d' $filepathPiperOSStep && \
    echo " - appended $filepathSAPSpecifics to $filepathPiperOSStep"
  done
else
  echo "No SAP-internal information found to append to piper-lib-os steps."
fi