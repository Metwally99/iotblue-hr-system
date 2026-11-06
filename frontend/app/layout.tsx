import type { Metadata } from 'next';
import { Geist, Geist_Mono } from 'next/font/google';
import './globals.css';
import './resources.css';
import './blue.css';
import { AuthProvider } from '@/components/auth-provider';

const geistSans = Geist({
  variable: '--font-geist-sans',
  subsets: ['latin'],
});

const geistMono = Geist_Mono({
  variable: '--font-geist-mono',
  subsets: ['latin'],
});

export const metadata: Metadata = {
  metadataBase: new URL(
    process.env.NEXT_PUBLIC_SITE_URL ?? 'http://localhost:3000',
  ),
  title: 'Humanly — HR Operations',
  description: 'A focused HR workspace for employees, leave, attendance, and reporting.',
  openGraph: {
    title: 'Humanly — HR Operations',
    description: 'Employees, leave, attendance, and reporting in one focused workspace.',
    images: ['/og.png'],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Humanly — HR Operations',
    description: 'Employees, leave, attendance, and reporting in one focused workspace.',
    images: ['/og.png'],
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        <AuthProvider>{children}</AuthProvider>
      </body>
    </html>
  );
}
