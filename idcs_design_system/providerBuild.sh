provider=$1

# Updating style.provider.scss
cat /dev/null > ./src/scss/providers/style.provider.scss
echo "@import \"$provider/index.scss\";" >> ./src/scss/providers/style.provider.scss
cat /dev/null > ./src/scss/providers/style.provider.dlux.scss
echo "@import \"$provider/_customize.dlux.lightTheme.scss\";" >> ./src/scss/providers/style.provider.dlux.scss
echo "@import \"$provider/_customize.dlux.darkTheme.scss\";" >> ./src/scss/providers/style.provider.dlux.scss

