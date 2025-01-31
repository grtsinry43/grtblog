import React from 'react';
import {Container} from '@radix-ui/themes';

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
