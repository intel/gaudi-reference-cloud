import {replaceInFile} from 'replace-in-file'
import fs from 'fs'
import {exit} from  'process'
import path from 'node:path'
import fse from 'fs-extra'
import configuration from './configuration.js'

const { copySync } = fse;

const environment = `${process.argv[2]}`

if (environment === 'undefined') {
    console.error('Missing argument[2] of environment.')
    exit(9)            
}

const config = configuration[environment]

if (config === 'undefined') {
    console.error(`Environment ${environment} not found on configuration.json`)
    exit(9)            
}


const policyName =  config.name;
const tenantId = config.tenantID;
const htmlTemplatesCDN = config.cdnHost;
const sourceDirectory = "./";
const distributionDirectoryName = "dist/policies";
const distributionDirectoryPath = `${sourceDirectory}${distributionDirectoryName}`;
const policiesDirectoryName = "Policies";
const policiesDirectoryPath = `${sourceDirectory}${policiesDirectoryName}`;

if (policyName === 'undefined') {
    console.error('Missing argument[2] of policy name.')
    exit(9)            
} else if (tenantId === 'undefined') {
    console.error('Missing argument[3] of tenant id.')
    exit(9)
}  else if (htmlTemplatesCDN === 'undefined') {
    console.error('Missing argument[4] of HTML templates cdn.')
    exit(9)
}

const ignoreFilesToCopy = [
  "README.md"
];

function removeDirectoryIfExist() {
    if (fs.existsSync(distributionDirectoryPath)) {
      fs.rmSync(distributionDirectoryPath, { recursive: true });
    }
}

function copyAndRenameFilesToDistributionFolder() {
    const copyFilter = (path) => {
      const shouldCopy =
        !ignoreFilesToCopy.some((x) => path.indexOf(x) > -1);
      return shouldCopy;
    };
  
    fs.readdirSync(policiesDirectoryPath, { withFileTypes: true }).filter((entry) => {
      const fullsrc = path.resolve(policiesDirectoryPath + path.sep + entry.name);
      const fulldest = path.resolve(
        distributionDirectoryPath + path.sep + entry.name
      );
      copySync(fullsrc, fulldest.replace('B2C_1A_NEW_', `B2C_1A_${policyName.toUpperCase()}_NEW_`), { filter: copyFilter });
      return true;
    });
}

async function doReplacements() {
    const options = {
        files: `${distributionDirectoryPath}/*`,
        from: [
            '/idcb2cdev.onmicrosoft.com/gi',
             '/B2C_1A_NEW_/gi',
            '/[$][$]CDN[$][$]\//gi',],
        to: [tenantId,
            `B2C_1A_${policyName.toUpperCase()}_NEW_`,
            `${htmlTemplatesCDN}/`
        ]
    }
    try {
        const results = await replaceInFile(options)
        console.log('Replacement results:', results)
     }
    catch (error) {
        console.error('Cannot replace values in files:', error)
        exit(1)
    }
}

async function createOrUpdateDistribution() {
    try {       
        console.log('Creating distribution for B2C XML files')
        removeDirectoryIfExist()
        copyAndRenameFilesToDistributionFolder()
        await doReplacements()
        exit(0)
    } catch (error) {
        console.log(error)
        exit(1)
    }
}

createOrUpdateDistribution()


