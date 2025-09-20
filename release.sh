#!/bin/bash

set -e

# Получение текущей версии
CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
echo "Current version: $CURRENT_VERSION"

# Извлечение номеров версии
VERSION_BITS=(${CURRENT_VERSION//v/ })
VERSION_BITS=(${VERSION_BITS//./ })
VNUM1=${VERSION_BITS[0]}
VNUM2=${VERSION_BITS[1]}
VNUM3=${VERSION_BITS[2]}

# Определение типа изменения по коммитам
COMMITS=$(git log ${CURRENT_VERSION}..HEAD --oneline)

if [[ $COMMITS == *"BREAKING CHANGE"* ]] || [[ $COMMITS == *"!"* ]]; then
    # Major version
    VNUM1=$((VNUM1+1))
    VNUM2=0
    VNUM3=0
elif [[ $COMMITS == *"feat:"* ]]; then
    # Minor version
    VNUM2=$((VNUM2+1))
    VNUM3=0
else
    # Patch version
    VNUM3=$((VNUM3+1))
fi

# Создание новой версии
NEW_TAG="v$VNUM1.$VNUM2.$VNUM3"
echo "New version: $NEW_TAG"

# Создание тега и пуш
git tag -a $NEW_TAG -m "Release $NEW_TAG"
git push origin $NEW_TAG

echo "Released $NEW_TAG"