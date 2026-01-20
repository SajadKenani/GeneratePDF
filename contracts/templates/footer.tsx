import { Head, Html, Preview, Text, Heading } from "@react-email/components"

import type { AlTaerEmailProps } from "../types"
import { logo, signature } from "../images"

export const FooterContractEmail = ({
  userName,
}: AlTaerEmailProps) => {
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
          <h1 className="m-0">3</h1>
        </div>
        <Preview>AL-TAER AL ASFAR - عقد بيع دجاج</Preview>
        <div className="mx-auto flex justify-center p-0">
          <div className="mx-4 mt-4 w-full max-w-4xl overflow-hidden rounded-2xl bg-white p-0 shadow-2xl">

            {/* Footer */}
            <section className="bg-white bg-gradient-to-br p-8 text-black">
              <div className="mb-8 rounded-xl border-2 border-emerald-400 bg-emerald-50 p-6 mt-10">
                <Heading className="mb-4 text-base font-bold text-gray-700">ملاحظات هامة</Heading>
                <Text className="m-0 text-sm leading-relaxed text-gray-700">
                  • يعتبر هذا العقد ساري المفعول من تاريخ التوقيع.
                  <br />
                  • أي تعديل على هذا العقد يجب أن يكون مكتوباً وموقعاً من الطرفين.
                  <br />
                  • في حالة النزاع، يتم الرجوع للقوانين المعمول بها في العراق.
                  <br />• الطرف الأول غير مسؤول عن النفوق الطبيعي أو الأمراض الوبائية.
                </Text>
              </div>
              <p className="m-0 mb-6 text-center text-base font-semibold">قررت وبكامل قناعتي قد وقعت العقد</p>
              <table className="w-full border-collapse">
                <tr>
                  <td className="w-1/3 p-4 text-center align-middle">
                    <p className="m-0 mb-2 text-sm font-semibold text-black">الطرف الأول</p>
                    <p className="m-0 mb-2 text-sm font-semibold text-black">شركة الطائر الأصفر</p>
                    <Text className="m-0 mb-2 text-sm font-bold text-gray-500">مهاب علي ياسين</Text>
                    <div className="mx-auto -mt-10 flex h-40 w-40 items-center justify-center">
                      <img src={`data:image/png;base64,${signature}`} width="128" height="128" alt="Signature" />
                    </div>
                  </td>

                  <td className="w-1/3 p-4 text-center align-middle">
                    <div className="mx-auto flex h-40 w-40 items-center justify-center">
                      <img src={`data:image/png;base64,${logo}`} width="128" height="128" alt="Logo" />
                    </div>
                  </td>

                  <td className="w-1/3 p-4 text-center align-middle">
                    <p className="m-0 mb-2 text-sm font-semibold text-black">الطرف الثاني المشتري</p>
                    <Text className="m-0 mb-2 text-sm font-bold text-gray-500">اسم المشتري: {userName}</Text>
                    <div className="mx-auto flex items-center justify-center bg-white" style={{ width: 64, height: 110 }}>
                      <hr className="mx-auto mt-4 w-full max-w-32 border-none bg-black" style={{ height: 1 }} />
                    </div>
                  </td>
                </tr>
              </table>
            </section>
          </div>
        </div>
      </body>
    </Html>
  )
}

export default FooterContractEmail
