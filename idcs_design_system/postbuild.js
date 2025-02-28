/**
 * Includes all package.json files to use subpath imports
 */

const fs = require('fs')
const path = require('path')
const packageJsonReference = require('./package.json')
const PackageJson = require('@npmcli/package-json')

function flatten(lists) {
  return lists.reduce((a, b) => a.concat(b), [])
}

function getDirectories(srcpath) {
  return fs
    .readdirSync(srcpath)
    .map((file) => path.join(srcpath, file))
    .filter((path) => fs.statSync(path).isDirectory())
}

function getDirectoriesRecursive(srcpath) {
  return [srcpath, ...flatten(getDirectories(srcpath).map(getDirectoriesRecursive))]
}

const directories = getDirectoriesRecursive('dist').map((x) => x.replaceAll('\\', '/'))

const getCjsPath = (dist) => {
  let relativePathToDist = ''
  for (let i = 0; i < dist.length; i++) {
    if (dist[i] === '/') {
      relativePathToDist = relativePathToDist + '../'
    }
  }
  return relativePathToDist === '' ? './' : relativePathToDist
}

const getPackageJsonExtra = (dir) => {
  if (dir === 'dist') {
    return {
      version: packageJsonReference.version,
      description: packageJsonReference.description,
      author: packageJsonReference.author,
      contributors: packageJsonReference.contributors,
      dependencies: packageJsonReference.dependencies,
      peerDependencies: packageJsonReference.peerDependencies
    }
  } else {
    return {
      main: `'${getCjsPath(dir)}index.cjs`,
      module: './index.js',
      types: './index.d.ts',
      private: true
    }
  }
}

const createPackageJson = async (dir) => {
  const packageJson = await PackageJson.create(dir)
  packageJson.update({
    name: `${packageJsonReference.name}${dir.replace('dist', '')}`,
    ...getPackageJsonExtra(dir)
  })
  await packageJson.save()
}

directories.forEach((dir) => {
  createPackageJson(dir)
})
