import { Head, Heading, Html } from "@react-email/components"

export const ContentContractEmail = ( content: string ) => {
    return (
        <Html dir="rtl" className="font-sans">
            <Head>
                <style>@import url('https://fonts.googleapis.com/css2?family=Tajawal:wght@200;300;400;500;700;800;900&display=swap');</style>
                <meta charSet="UTF-8" />
                <meta name="viewport" content="width=device-width, initial-scale=1.0" />
                <title>AL-TAER AL ASFAR - Business Document</title>
            </Head>

            <script src="https://cdn.tailwindcss.com"></script>
            <body className="m-0 bg-gray-50 p-0 font-sans" style={{ fontFamily: "'Tajawal', sans-serif" }}>
                <div
                    className="fixed top-0 left-0 z-1000 flex items-center justify-center text-white"
                    style={{
                        width: 30,
                        height: 30,
                        margin: 6,
                        backgroundColor: "black",
                        borderRadius: "50%",
                        zIndex: 1000,
                    }}
                >
                    <h1 className="m-0">2</h1>
                </div>
                <div className="mx-auto flex justify-center p-0">
                    <div className="mx-4 mt-4 w-full max-w-4xl overflow-hidden rounded-2xl bg-white p-0 shadow-2xl">
                        <div className="p-8">
                            <div className="mb-8">
                                <Heading className="m-0 text-lg font-bold text-gray-800">أحكام العقد</Heading>
                                <div id="content" dangerouslySetInnerHTML={{ __html: content }} />
                            </div>
                        </div>
                    </div>
                </div>

            </body>
        </Html>
    )
}

export default ContentContractEmail
