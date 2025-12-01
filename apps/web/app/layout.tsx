import type { Metadata } from "next";
import "./globals.css";
import { AuthProvider } from "@/lib/auth_context";

export const metadata: Metadata = {
    title: "Arc Privacy Platform",
    description: "Enterprise-grade consent management and privacy compliance platform",
    keywords: ["privacy", "GDPR", "DPDP", "consent management", "data protection"],
};

export default function RootLayout({
    children,
}: Readonly<{
    children: React.ReactNode;
}>) {
    return (
        <html lang="en">
            <body>
                <AuthProvider>
                    {children}
                </AuthProvider>
            </body>
        </html>
    );
}
