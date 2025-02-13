import type {Metadata} from 'next';
import '@/styles/global.css';
import './globals.css';
import {Theme} from '@radix-ui/themes';
import '@radix-ui/themes/styles.css';
import {ThemeProvider} from './theme-provider';
import styles from '@/styles/Main.module.scss';
import NavBar from '@/components/navbar/NavBar';
import React from 'react';
import Footer from '@/components/footer/Footer';
import StoreProvider from '@/app/store-provider';
import {getAllNavItem} from '@/api/navbar';
import {getWebsiteInfo} from "@/api/websiteInfo";
import {WebsiteInfo} from "@/types";
import {WebsiteInfoProvider} from "@/app/website-info-provider";
import {OpenPanelComponent} from "@openpanel/nextjs";
import GlobalNotification from "@/components/notification/GlobalNotification";
import UpdateNotification from "@/components/notification/UpdateNotification";
import {ToastContainer} from "react-toastify";
import 'react-toastify/dist/ReactToastify.css';
import {NotificationProvider} from "@/lib/NotificationContext";

const websiteInfo: WebsiteInfo = await getWebsiteInfo({next: {revalidate: 60}});

export const metadata: Metadata = {
    title: websiteInfo.WEBSITE_NAME,
    description: websiteInfo.WEBSITE_DESCRIPTION,
    keywords: websiteInfo.WEBSITE_KEYWORDS,
};

export default async function RootLayout({
                                             children,
                                         }: Readonly<{
    children: React.ReactNode;
}>) {
    const navItems = await getAllNavItem({next: {revalidate: 60}});
    return (
        <html lang="en" suppressHydrationWarning>
            <body>
                <OpenPanelComponent
                    clientId="c03c087d-b51c-4263-9659-ab9277b2f09c"
                    trackScreenViews={true}
                    // trackAttributes={true}
                    // trackOutgoingLinks={true}
                    // If you have a user id, you can pass it here to identify the user
                    // profileId={'123'}
                />
                <StoreProvider>
                    <ThemeProvider
                        attribute="class"
                        enableSystem
                        disableTransitionOnChange={false}
                    >
                        <Theme>
                            <WebsiteInfoProvider websiteInfo={websiteInfo}>
                                <NotificationProvider>
                                    <NavBar items={navItems}/>
                                    <div className={styles.mainContainer}>
                                        {children}
                                    </div>
                                    <Footer websiteInfo={websiteInfo}/>
                                </NotificationProvider>
                            </WebsiteInfoProvider>
                            <GlobalNotification/>
                            <UpdateNotification/>
                            <ToastContainer position="top-center"/>
                        </Theme>
                    </ThemeProvider>
                </StoreProvider>
            </body>
        </html>
    );
}
