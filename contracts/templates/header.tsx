// import { Head, Heading, Html, Img, Preview, Text } from "@react-email/components"

// import type { AlTaerEmailProps } from "../types"
// import { mainLogo } from "../images"

// export const HeaderContractEmail = ({ commonProps, title }: {commonProps: AlTaerEmailProps, title: string}) => {
//     const date = commonProps.createdAt.toLocaleDateString("en-IQ", {
//         day: "2-digit",
//         month: "numeric",
//         year: "numeric",
//     })

//     return (
//         <Html dir="rtl" className="font-sans">
//             <Head>
//                 <style>@import url('https://fonts.googleapis.com/css2?family=Tajawal:wght@200;300;400;500;700;800;900&display=swap');</style>
//                 <meta charSet="UTF-8" />
//                 <meta name="viewport" content="width=device-width, initial-scale=1.0" />
//                 <title>AL-TAER AL ASFAR - Business Document</title>
//             </Head>

//             <script src="https://cdn.tailwindcss.com"></script>
//             <body className="m-0 bg-gray-50 p-0 font-sans" style={{ fontFamily: "'Tajawal', sans-serif" }}>
//                 <div
//                     className="fixed top-0 left-0 z-1000 flex items-center justify-center text-white"
//                     style={{
//                         width: 30,
//                         height: 30,
//                         margin: 6,
//                         backgroundColor: "black",
//                         borderRadius: "50%",
//                         zIndex: 1000,
//                     }}
//                 >
//                     <h1 className="m-0">1</h1>
//                 </div>
//                 <Preview>AL-TAER AL ASFAR - عقد بيع دجاج</Preview>
//                 <div className="mx-auto flex justify-center p-0">
//                     <div className="mx-4 mt-4 w-full max-w-4xl overflow-hidden rounded-2xl bg-white p-0 shadow-2xl">
//                         {/* Header */}
//                         <div className="relative bg-gradient-to-br from-indigo-500 to-purple-600 p-8 text-white">
//                             <table className="mb-5 w-full">
//                                 <tr>
//                                     <td className="w-48 align-middle">
//                                         <div className="flex items-center gap-1">
//                                             <div className="inline-block h-6 w-6 rounded-full bg-yellow-400"></div>
//                                             <div className="inline-block h-8 w-8 rounded-full bg-gray-500"></div>
//                                             <div className="inline-block h-10 w-10 rounded-full bg-gray-700"></div>
//                                             <div className="ml-3 inline-block h-8 w-20 rounded-2xl bg-gradient-to-r from-yellow-400 to-orange-400"></div>
//                                         </div>
//                                     </td>
//                                     <td className="text-right align-middle">
//                                         <Heading className="m-0 mb-2 text-3xl font-bold text-white">شركة الطائر الأصفر</Heading>
//                                         <Text className="m-0 mb-2 text-base text-white opacity-90">AL-TAER AL ASFAR</Text>
//                                         <Text className="m-0 text-sm leading-relaxed text-white opacity-90">
//                                             العراق - بغداد - شارع الربيعي - عمارة سنتر الربيعي
//                                             <br />
//                                             مقابل مول جوهرة بغداد
//                                         </Text>
//                                     </td>
//                                     <td className="text-right align-middle">
//                                         <Img src={`data:image/png;base64,${mainLogo}`} width="128" height="128" alt="Logo" />
//                                     </td>
//                                 </tr>
//                             </table>

//                             <div className="mt-5 rounded-lg bg-gradient-to-r from-gray-600 to-gray-800 p-4">
//                                 <Heading className="m-0 text-center text-xl font-bold text-white">{title}</Heading>
//                             </div>
//                         </div>

//                         {/* Content */}
//                         <div className="p-8">
//                             {/* Contract Parties */}
//                             <div className="mb-6 rounded-lg border border-slate-200 bg-slate-50 p-5">
//                                 <Heading className="m-0 mb-4 text-lg font-bold text-gray-800">أطراف العقد</Heading>
//                                 <Text className="m-0 text-sm leading-relaxed text-gray-600">
//                                     <strong>1. البائع (الطرف الأول):</strong> شركة الطائر الأصفر ومقرها في بغداد - شارع الربيعي - زينوبة، مقابل مول جوهرة بغداد. ويشار إليه
//                                     فيما يلي بالطرف الأول ويمثلها في هذا العقد السيد: مهاب علي ياسين
//                                 </Text>
//                             </div>

//                             {/* Customer Details */}
//                             <div className="mb-8">
//                                 <Heading className="m-0 mb-4 text-lg font-bold text-gray-800">بيانات المشتري</Heading>

//                                 <table className="mb-2 w-full">
//                                     <tbody>
//                                         <tr>
//                                             <td className="w-1/4 p-2 align-top">
//                                                 <Text className="m-0 mb-2 text-sm font-bold text-gray-500">اسم المشتري:</Text>
//                                                 <div className="rounded-lg border-2 border-dashed border-gray-300 bg-gray-50 px-4 py-2 text-sm font-bold text-red-600">{commonProps.userName}</div>
//                                             </td>

//                                             <td className="w-1/4 p-2 align-top">
//                                                 <Text className="m-0 mb-2 text-sm font-bold text-gray-500">Invoice</Text>
//                                                 <div className="rounded-lg border-2 border-dashed border-gray-300 bg-gray-50 px-4 py-2 text-sm font-bold text-red-600">
//                                                     {commonProps.depositNumId}
//                                                 </div>
//                                             </td>

//                                             <td className="w-1/4 p-2 align-top">
//                                                 <Text className="m-0 mb-2 text-sm font-bold text-gray-500">رقم الهاتف:</Text>
//                                                 <div className="rounded-lg border-2 border-dashed border-gray-300 bg-gray-50 px-4 py-2 text-sm font-bold text-red-600">{commonProps.userPhone}</div>
//                                             </td>
//                                         </tr>

//                                         <tr>
//                                             <td className="w-1/2 p-2 align-top">
//                                                 <Text className="m-0 mb-2 text-sm font-bold text-gray-500">البريد الإلكتروني:</Text>
//                                                 <div className="rounded-lg border-2 border-dashed border-gray-300 bg-gray-50 px-4 py-2 text-sm font-bold text-red-600">{commonProps.userEmail}</div>
//                                             </td>

//                                             <td className="w-1/2 p-2 align-top">
//                                                 <Text className="m-0 mb-2 text-sm font-bold text-gray-500">Date</Text>
//                                                 <div className="rounded-lg border-2 border-dashed border-gray-300 bg-gray-50 px-4 py-2 text-sm font-bold text-red-600">{date}</div>
//                                             </td>
//                                         </tr>
//                                     </tbody>
//                                 </table>
//                             </div>

//                             {/* Contract Details */}
//                             <div className="mb-8 overflow-hidden rounded-xl border-2 border-gray-200 bg-gray-50">
//                                 <table className="w-full border-collapse">
//                                     <tr>
//                                         <td className="w-3/10 border-l border-gray-200 p-6 align-top">
//                                             <Text className="m-0 mb-3 text-sm leading-relaxed text-gray-700">
//                                                 نوع الباقة: <span className="font-bold text-red-600">{commonProps.item}</span>
//                                             </Text>
//                                             <Text className="m-0 mb-3 text-sm leading-relaxed text-gray-700">
//                                                 العدد: <span className="font-bold text-red-600">{commonProps.itemAmount}</span>
//                                             </Text>
//                                         </td>
//                                         <td className="w-7/10 p-6 align-top">
//                                             <Text className="m-0 mb-3 text-sm leading-relaxed text-gray-700">
//                                                 المبلغ الكلي المستلم:{" "}
//                                                 <strong>
//                                                     {commonProps.currencyAmount}&nbsp;{commonProps.currency.toUpperCase()}
//                                                 </strong>{" "}
//                                             </Text>
//                                             <Text className="m-0 mb-3 text-sm leading-relaxed text-gray-700">
//                                                 عملية الدفع: <span className="font-bold text-red-600">{commonProps.paymentMethod}</span>
//                                             </Text>
//                                         </td>
//                                         <td className="w-3/10 border-l border-gray-200 p-6 align-top">
//                                             <Text className="m-0 mb-3 text-sm leading-relaxed text-gray-700">مؤمن عليه: {commonProps.insurance === true ? "نعم" : "لا"}</Text>
//                                         </td>
//                                     </tr>
//                                 </table>
//                             </div>
//                         </div>

//                     </div>
//                 </div>
//             </body>
//         </Html>
//     )
// }

// export default HeaderContractEmail
