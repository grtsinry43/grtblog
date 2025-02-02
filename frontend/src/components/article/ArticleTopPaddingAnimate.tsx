'use client';

import React from 'react';
import {motion} from 'framer-motion';

const spring = {
    type: 'spring',
    stiffness: 300,
    damping: 10,
    mass: 1,
    bounce: 0.5,
};

const ArticleTopPaddingAnimate = ({reverse = false, once = false}: { reverse?: boolean, once?: boolean }) => {
    return (
        <motion.div
            initial={{paddingTop: reverse ? 50 : 0}}
            animate={{paddingTop: reverse ? 0 : 50}}
            transition={once ? {type: 'spring', stiffness: 300, damping: 10, mass: 1, bounce: 0.5} : spring}
            style={{width: '100%'}}
        >
            {/* Content goes here */}
        </motion.div>
    );
};

export default ArticleTopPaddingAnimate;
