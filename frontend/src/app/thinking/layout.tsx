import type {Metadata} from 'next';
import React from 'react';
import {Container} from '@radix-ui/themes';

export const metadata: Metadata = {
    title: '思考',
    description: '这是我的所思所想，与你分享。',
};

export default function RootLayout({
                                       children,
                                   }: Readonly<{
    children: React.ReactNode;
}>) {
    return (
        <Container size={'4'} style={{
            padding: '0 20px',
            scrollBehavior: 'smooth',
        }}>
            {children}
        </Container>
    );
}
