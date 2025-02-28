@ECHO OFF

REM Command file for Sphinx documentation

if "%SPHINXBUILD%" == "" (
	set SPHINXBUILD=sphinx-build
)

set SCRIPTDIR=source\_scripts
set BUILDDIR=source\_build
set ALLSPHINXOPTS=-d %BUILDDIR%/doctrees %SPHINXOPTS% source


if "%1" == "" goto help

if "%1" == "help" (
	:help
	echo.Please use `make ^<target^>` where ^<target^> is one of
	echo.  html       to make standalone HTML files
	echo.  clean	  to remove all files from source/_build directory  
	echo.  tut		  to add a tutorial using tutorial template
	echo.  tutdemo    to walk through creating a tutorial from a template
 	echo.  tutclean   to clean up after tutorial demo, or remote tutorial template
	goto end
)

if "%1" == "clean" (
	for /d %%i in (%BUILDDIR%\*) do rmdir /q /s %%i
	del /q /s %BUILDDIR%\*
	echo.The %BUILDDIR%\* was cleaned. 
	goto end
)


REM Check if sphinx-build is available and fallback to Python version if any
%SPHINXBUILD% 2> nul
if errorlevel 9009 goto sphinx_python
goto sphinx_ok

:sphinx_python

set SPHINXBUILD=python -m sphinx.__init__
%SPHINXBUILD% 2> nul
if errorlevel 9009 (
	echo.
	echo.The 'sphinx-build' command was not found. Make sure you have Sphinx
	echo.installed, then set the SPHINXBUILD environment variable to point
	echo.to the full path of the 'sphinx-build' executable. Alternatively you
	echo.may add the Sphinx directory to PATH.
	echo.
	echo.If you don't have Sphinx installed, grab it from
	echo.http://sphinx-doc.org/
	exit /b 1
)

:sphinx_ok

if "%1" == "html" (
	%SPHINXBUILD% -b html %ALLSPHINXOPTS% %BUILDDIR%/html
	if errorlevel 1 exit /b 1
	echo.
	echo.Build finished. The HTML pages are in %BUILDDIR%/html.
	goto end
)

if "%1" == "tut" (
	echo.%SCRIPTDIR%
	cd %SCRIPTDIR%
	python.exe tut.py
	echo.
	goto end
)

if "%1" == "tutdemo" (
	cd %SCRIPTDIR%
	python.exe tut_demo.py
	echo.
	goto end
)

if "%1" == "tutclean" (
	cd %SCRIPTDIR%
	python.exe cln_demo.py
	echo.
	goto end
)

%SPHINXBUILD% -M %1 %SOURCEDIR% %BUILDDIR% %SPHINXOPTS% %O%
goto end

:help
%SPHINXBUILD% -M help %SOURCEDIR% %BUILDDIR% %SPHINXOPTS% %O%

:end

