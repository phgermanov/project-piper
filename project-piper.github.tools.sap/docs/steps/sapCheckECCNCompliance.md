# ${docGenStepName}

## ${docGenDescription}

## ${docGenParameters}

## Configuration

In order to enable the ECCN check for your pipeline, the project needs to have a PPMS object number of Software Component Version (SCV) or Product Version (PV). Furthermore, the IFP credentials need to be added to your Jenkins server and the .pipeline/config.yml has to be adjusted as shown below:

``` YAML
steps:
  sapCheckECCNCompliance:
    eccnCredentialsId: 'Enter IFP credential ID from jenkins'
    ppmsID: 'Enter PPMS object ID here'
```

## Prerequisites

This pipeline step can only be executed successfully if you have done the following activities in advance:

### Credentials for ECCN from IFP system

Create a credential entry (with type username and password) in your Jenkins for ECCN check with your IFP username and password. If you are unsure about how to create credential entries in Jenkins, you can find further information in the [Jenkins documentation](https://jenkins.io/doc/book/using/using-credentials/).

**This ensures that your password is stored securely in your Jenkins system.**

<a name="hint"> </a>
!!! hint "Hint: How to get a password for IFP Technical user"
    You need to have the credentials of the technical user in IFP system to connect to ECCN Manager via REST. In case you don’t have a technical user already, create an [IT Direct](https://itdirect.wdf.sap.corp/sap(bD1lbiZjPTAwMSZkPW1pbg==)/bc/bsp/sap/crm_ui_start/default.htm?crm-object-type=AIC_OB_INCIDENT&crm-object-action=D&PROCESS_TYPE=ZINE&CAT_ID=SRAS_IAM_AUTH) ticket (Category ID: **SRAS_IAM_AUTH**) to create a technical user in IFP. Provide a technical username in the format **RFC_ECCN_XXX** (max up to 12 characters) and request for **0000_IFP_CP_ECCN_API_READ** role to be assigned to the technical user in IFP. Also, share your passvault user group to receive access to the password in the passvault. In case you already have the technical user, and you don’t remember the password, create an [IT Direct](https://itdirect.wdf.sap.corp/sap(bD1lbiZjPTAwMSZkPW1pbg==)/bc/bsp/sap/crm_ui_start/default.htm?crm-object-type=AIC_OB_INCIDENT&crm-object-action=D&PROCESS_TYPE=ZINE&CAT_ID=SRAS_IAM_AUTH) ticket (Category ID: **SRAS_IAM_AUTH**) to reset the password for the technical user in IFP. Share your passvault user group to receive access to the password in the passvault. After you have done this, you can use technical user and the password in your Jenkins.
