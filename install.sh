set -e

VERSION=0.4.0-beta.22

if command -v apk; then
  EXT=apk
  INSTALL="apk add"
elif command -v yum; then
  EXT=yum
  INSTALL="yum install"
elif command -v rpm; then
  EXT=rpm
  INSTALL="rpm install"
elif command -v apt-get; then
  EXT=deb
  INSTALL="apt-get install"
else
  echo "Unsupported package manager"
  exit 1
fi

LOC="https://github.com/elseano/rundown/releases/download/v${VERSION}/rundown_${VERSION}_linux_amd64.$EXT"
echo "Downloading $LOC..."

if command -v curl; then
  curl -L -O -k $LOC
elif command -v wget; then
  wget $LOC
else
  echo "Need wget or curl installed"
  exit 1
fi

$INSTALL ./rundown_${VERSION}_linux_amd64.$EXT
rm rundown_${VERSION}_linux_amd64.$EXT