@echo off
echo Committing without linting checks...
git commit --no-verify %*
echo Done!
