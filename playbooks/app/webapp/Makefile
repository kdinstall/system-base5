.PHONY: build watch run binary dev clean

## CSS ビルド（Tailwind v4）
build:
	npm run build

## CSS ウォッチ（バックグラウンド推奨）
watch:
	npm run watch

## Go アプリ起動（go run）
run: build
	go run ./src

## Go バイナリのビルド（CSS ビルドも含む）
binary: build
	go build -o webapp ./src

## 開発用: go mod tidy + npm install + CSS ビルド + 起動
dev:
	go mod tidy
	npm install
	npm run build
	go run ./src

## 生成物の削除
clean:
	rm -f webapp webapp.exe
	rm -f public/assets/css/style.css
