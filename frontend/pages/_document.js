import { Html, Head, Main, NextScript } from "next/document";

export default function Document() {
    return (
        <Html lang="id">
            <Head>
                <link rel="icon" type="image/png" href="/favicon.png" />
                <link rel="apple-touch-icon" href="/apple-touch-icon.png" />
            </Head>
            <body>
                <Main />
                <NextScript />
            </body>
        </Html>
    );
}
