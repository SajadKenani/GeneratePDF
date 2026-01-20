import { Head, Heading, Html, Img, Preview, Text } from "@react-email/components"

import type { AlTaerEmailProps } from "../../types"
import { logo, mainLogo } from "../../images"

export const AlTaerReceiptEmail = ({
    depositNumId,
    createdAt,
    item,
    itemAmount,
    currency,
    currencyAmount,
    paymentMethod
}: AlTaerEmailProps) => {
    const date = createdAt.toLocaleDateString("ar-IQ", {
        day: "2-digit",
        month: "long",
        year: "numeric",
    })

    const time = createdAt.toLocaleTimeString("ar-IQ", {
        hour: "2-digit",
        minute: "2-digit",
    })

    return (
        <Html dir="rtl" className="font-sans">
            <Head>
                <style>@import url('https://fonts.googleapis.com/css2?family=Tajawal:wght@200;300;400;500;700;800;900&display=swap');</style>
                <meta charSet="UTF-8" />
                <meta name="viewport" content="width=device-width, initial-scale=1.0" />
                <title>AL-TAER AL ASFAR - وصل استلام</title>
            </Head>

            <script src="https://cdn.tailwindcss.com"></script>
            <Preview>AL-TAER AL ASFAR - وصل استلام رقم {String(depositNumId)}</Preview>

            <body className="m-0 bg-gray-50 p-5 font-sans" style={{ fontFamily: "'Tajawal', sans-serif" }}>
                <div className="mx-auto max-w-2xl overflow-hidden rounded-xl bg-white shadow-lg">
                    {/* Header */}
                    <div className="bg-gradient-to-br from-indigo-600 to-purple-700 p-6 text-white">
                        <table className="w-full">
                            <tbody>
                                <tr>
                                    <td className="w-24 align-middle">
                                        <Img
                                            src={`data:image/png;base64,${mainLogo}`}
                                            width="96"
                                            height="96"
                                            alt="Logo"
                                            className="block rounded-lg"
                                        />
                                    </td>

                                    <td className="pr-4 text-right align-middle">
                                        <Heading className="m-0 mb-1 text-2xl font-bold text-white">
                                            شركة الطائر الأصفر
                                        </Heading>
                                        <Text className="m-0 mb-1 text-sm font-semibold text-white opacity-90">
                                            AL-TAER AL ASFAR
                                        </Text>
                                        <Text className="m-0 text-xs leading-relaxed text-white opacity-80">
                                            العراق - بغداد - شارع الربيعي
                                        </Text>
                                    </td>
                                </tr>
                            </tbody>
                        </table>

                        <div className="mt-4 rounded-lg bg-white bg-opacity-20 p-3 text-center backdrop-blur-sm">
                            <Heading className="m-0 text-xl font-bold text-white">وصل استلام</Heading>
                            <Text className="m-0 mt-1 text-sm text-white opacity-90">Payment Receipt</Text>
                        </div>
                    </div>

                    {/* Three-row summary: ID, package+count, price */}
                    <div className="border-b-2 border-dashed border-gray-300 bg-gradient-to-r from-gray-50 to-white p-6">
                        <table className="w-full">
                            <tbody>
                                <tr>
                                    <td className="pb-2 text-right">
                                        <Text className="m-0 text-sm font-semibold text-gray-600">رقم الوصل:</Text>
                                    </td>
                                    <td className="pb-2 text-left">
                                        <Text className="m-0 text-sm font-bold text-indigo-600">#{depositNumId}</Text>
                                    </td>
                                </tr>
                                <tr>
                                    <td className="pb-2 text-right">
                                        <Text className="m-0 text-sm font-semibold text-gray-600">الباقة وعدد الدجاج:</Text>
                                    </td>
                                    <td className="pb-2 text-left">
                                        <Text className="m-0 text-sm font-semibold text-gray-800">{item} × {itemAmount}</Text>
                                    </td>
                                </tr>
                                <tr>
                                    <td className="text-right">
                                        <Text className="m-0 text-sm font-semibold text-gray-600">السعر:</Text>
                                    </td>
                                    <td className="text-left">
                                        <Text className="m-0 text-xl font-bold text-indigo-600">{currencyAmount.toLocaleString()} {currency}</Text>
                                    </td>
                                </tr>
                                   <tr>
                                    <td className="text-right">
                                        <Text className="m-0 text-sm font-semibold text-gray-600">طريقة الدفع:</Text>
                                    </td>
                                    <td className="text-left">
                                        <Text className="m-0 text-xl font-bold text-indigo-600">{paymentMethod}</Text>
                                    </td>
                                </tr>
                            </tbody>
                        </table>
                        <div className="mt-3 text-center">
                            <Text className="m-0 text-xs text-gray-500">التاريخ: {date} • الوقت: {time}</Text>
                        </div>
                    </div>

                    {/* Footer */}
                    <div className="bg-gradient-to-r from-gray-50 to-white p-6 text-center">
                        <div className="mb-4 flex items-center justify-center gap-2">
                            <Img
                                src={`data:image/png;base64,${logo}`}
                                width="48"
                                height="48"
                                alt="Logo"
                                className="block opacity-60"
                            />
                        </div>

                        <Text className="m-0 mb-2 text-sm font-semibold text-gray-700">
                            شكراً لتعاملكم معنا
                        </Text>
                        <Text className="m-0 mb-3 text-xs text-gray-500">
                            Thank you
                        </Text>

                        <div className="mx-auto mb-3 h-px w-24 bg-gradient-to-r from-transparent via-gray-300 to-transparent"></div>

                        <Text className="m-0 text-xs leading-relaxed text-gray-500">
                            للاستفسارات: www.newyolkfarm.com | 07717688884
                            <br />
                            العراق - بغداد - شارع الربيعي - عمارة سنتر الربيعي
                        </Text>

                        <Text className="m-0 mt-3 text-xs italic text-gray-400">
                            هذا الوصل تم إنشاؤه إلكترونياً ولا يحتاج إلى ختم أو توقيع
                        </Text>
                    </div>
                </div>
            </body>
        </Html>
    )
}

export default AlTaerReceiptEmail