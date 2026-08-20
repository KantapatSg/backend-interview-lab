import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Backend Interview Lab",
  description: "CQRS, Kafka and job workflow dashboard",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
