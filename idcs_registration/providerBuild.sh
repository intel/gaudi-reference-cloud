provider=$1

# Updating style.provider.scss
cat /dev/null > ../idcs_design_system/src/scss/providers/style.provider.scss
echo "@import \"$provider/index.scss\";" >> ../idcs_design_system/src/scss/providers/style.provider.scss
cat /dev/null > ../idcs_design_system/src/scss/providers/style.provider.dlux.scss
echo "@import \"$provider/_customize.dlux.lightTheme.scss\";" >> ../idcs_design_system/src/scss/providers/style.provider.dlux.scss
echo "@import \"$provider/_customize.dlux.darkTheme.scss\";" >> ../idcs_design_system/src/scss/providers/style.provider.dlux.scss

