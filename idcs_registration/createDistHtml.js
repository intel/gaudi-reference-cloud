
import {replaceInFile} from 'replace-in-file'
import fs from 'fs'
import { exit } from 'process'
import sass from 'sass'
import path from 'node:path'
import fse from 'fs-extra'
import {minify} from 'minify';
import tryToCatch from 'try-to-catch';
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

const htmlTemplatesCDN =  config.cdnHost
const signUpURL = config.signupURL
const consoleName = config.consoleName

const ignoreFilesToCopy = [
    "README.md",
    "*.scss"
  ];

const sourceDirectory = "./";
const distributionDirectoryName = "dist/html/";
const distributionDirectoryPath = `${sourceDirectory}${distributionDirectoryName}`;
const htmlDirectoryName = "Html";
const htmlDirectoryPath = `${sourceDirectory}${htmlDirectoryName}`;
const cssDirectoryName = "static/css";
const cssDirectoryPath = `${distributionDirectoryName}${cssDirectoryName}`;
const cssFileName = `main.${Date.now()}.css`

function copyFilesToDistributionFolder() {
    const copyFilter = (path) => {
      const shouldCopy =
        !ignoreFilesToCopy.some((x) => {
            const isWildcard = x.indexOf('*') !== -1
            if (isWildcard) {
                return path.endsWith(x.replace('*', ''))
            }
            return path.indexOf(x) > -1
        } );
      return shouldCopy;
    };
  
    fs.readdirSync(htmlDirectoryPath, { withFileTypes: true }).filter((entry) => {
      const fullsrc = path.resolve(htmlDirectoryPath + path.sep + entry.name);
      const fulldest = path.resolve(
        distributionDirectoryPath + path.sep + entry.name
      );
      copySync(fullsrc, fulldest, { filter: copyFilter });
      return true;
    });

    fs.mkdirSync(`${distributionDirectoryPath}/static/css`)
}
  
function removeDirectoryIfExist() {
    if (fs.existsSync(distributionDirectoryPath)) {
    fs.rmSync(distributionDirectoryPath, { recursive: true });
    }
}

async function doReplacementsCss() {
    const options = {
        files: `${cssDirectoryPath}/*`,
        from: [
            '/..\/..\/..\/..\/src\/assets\/images\//i',
             '/..\/fonts\//i'
        ],
        to: [
            `${htmlTemplatesCDN}/static/media/`,
            `${htmlTemplatesCDN}/static/fonts/`
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

async function doReplacementsHtml() {
    const options = {
        files: `${distributionDirectoryPath}/*`,
        from: [
            '/[$][$]CDN[$][$]\//gi',
            '/[$][$]SIGNUPURL[$][$]/gi',
            '/[$][$]CONSOLENAME[$][$]/gi',
            '/main.css'
        ],
        to: [
            `${htmlTemplatesCDN}/`,
            signUpURL,
            consoleName,
            `/${cssFileName}`
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

async function minifyHtml() {
    const filesToProccess = []
    fs.readdirSync(distributionDirectoryPath, { withFileTypes: true }).filter((entry) => {
        const fullsrc = path.resolve(distributionDirectoryPath + path.sep + entry.name);
        if (fullsrc.endsWith('.html')) {
            filesToProccess.push(fullsrc)
        }
        return true;
    });
    const options = {
        html: {
          removeComments: true,
          removeCommentsFromCDATA: true,
          removeCDATASectionsFromCDATA: true,
          collapseWhitespace: false,
          collapseBooleanAttributes: true,
          removeAttributeQuotes: true,
          removeRedundantAttributes: true,
          useShortDoctype: true,
          removeEmptyAttributes: false,
          removeEmptyElements: false,
          removeOptionalTags: false,
          removeScriptTypeAttributes: true,
          removeStyleLinkTypeAttributes: true,
          minifyJS: true,
          minifyCSS: true
        }
      }
      async function minifyHtml(src) {
        const [error, data] = await tryToCatch(minify, src, options);
        if (error)
            return console.error(error.message);
        return {data, src}
      }
      const promises = filesToProccess.map(x => minifyHtml(x))
      const results = await Promise.all(promises)
      results.forEach(minifiedFile => {
        fs.writeFileSync(minifiedFile.src, minifiedFile.data, (err) => {
            if (err) {
                console.error(`Cannot replace minified content of file ${minifiedFile.src}`)
            }
        })
      });
}

function compileSass() {
    const result = sass.compile(`./Html/app.scss`, {
        style: "compressed",
    })
    
    fs.writeFile(`${distributionDirectoryPath}/static/css/${cssFileName}`, result.css, (err)=>{
        if(err !== null) {
            console.error(`Cannot create css file. ${err.message}`);
            exit(err.code)
        }        
    })
}

async function createOrUpdateDistribution() {
    try {       
        console.log('Creating distribution for HTML files')
        removeDirectoryIfExist()
        copyFilesToDistributionFolder()
        compileSass()
        await doReplacementsCss()
        await doReplacementsHtml()
        await minifyHtml()
        exit(0)
    } catch (error) {
        console.log(error)
        exit(1)
    }
}

createOrUpdateDistribution()