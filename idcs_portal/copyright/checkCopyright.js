const fs = require('fs')
const path = require('path')

function readHeaderFromFile(callback) {
  const filePath = './copyright/copyrightHeader.txt'
  fs.readFile(filePath, 'utf8', (err, data) => {
    if (err) {
      console.error('Error reading txt:', err)
      return
    }
    callback(data.trim().split('\n'))
  })
}

function processFiles(folderPath, textLines) {
  fs.readdir(folderPath, (err, files) => {
    if (err) {
      console.error('Error reading directory:', err)
      return
    }

    files.forEach((file) => {
      const filePath = path.join(folderPath, file)

      fs.stat(filePath, (err, stats) => {
        if (err) {
          console.error('Error getting file stats:', err)
          return
        }

        if (stats.isDirectory()) {
          processFiles(filePath, textLines)
        } else {
          const fileExt = path.extname(filePath).toLowerCase()
          if (['.js', '.ts', '.jsx', '.tsx'].includes(fileExt)) {
            fs.readFile(filePath, 'utf8', (err, data) => {
              if (err) {
                console.error('Error reading file:', err)
                return
              }

              if (!textLines.some((line) => data.includes(line))) {
                const newData = `${textLines.join('\n')}\n\n${data}`
                fs.writeFile(filePath, newData, (err) => {
                  if (err) {
                    console.error('Error writing to file:', err)
                    return
                  }
                  console.log(`Added copyright to ${filePath}. Include this file on future commits.`)
                })
              }
            })
          }
        }
      })
    })
  })
}

// Call
const folderPath = './src'
readHeaderFromFile((textLines) => {
  processFiles(folderPath, textLines)
})
