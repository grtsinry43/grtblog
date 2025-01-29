"use client";

import React, {useEffect} from 'react';
import {reportRead} from "@/api/read-report";

const UserReadReporter = ({articleId}: { articleId: string }) => {
    useEffect(() => {
        const timer = setInterval(async () => {
            await reportRead(articleId);
        }, 5000);
        return () => clearInterval(timer);
    }, [articleId]);
    return (
        <>
        </>
    );
};

export default UserReadReporter;
