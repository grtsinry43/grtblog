'use client';

import React, {useEffect, useState} from 'react';
import {AnimatePresence, motion} from 'framer-motion';
import {Button, IconButton, TextField, Tooltip} from '@radix-ui/themes';
import styles from '@/styles/LoginModal.module.scss';
import {CloseIcon} from 'next/dist/client/components/react-dev-overlay/internal/icons/CloseIcon';
import {clsx} from 'clsx';
import {varela_round, noto_sans_sc} from '@/app/fonts/font';
import {MailIcon} from 'lucide-react';
import {GitHubLogoIcon} from '@radix-ui/react-icons';
import {FaGoogle} from 'react-icons/fa';
// import {IoLogoWechat} from 'react-icons/io5';
// import {BiLogoMicrosoft} from 'react-icons/bi';
import {userLogin, userRegister} from '@/api/user';
import {UserInfo} from '@/redux/userSlice';
import {useAppDispatch} from '@/redux/hooks';
import Link from "next/link";
import {toast} from "react-toastify";
import {Separator} from '@/components/ui/separator';

const LoginModal = ({isOpen, onClose}: { isOpen: boolean; onClose: () => void }) => {
    const [loginForm, setLoginForm] = useState({
        userEmail: '',
        password: '',
    });
    const [registerForm, setRegisterForm] = useState({
        nickname: '',
        userEmail: '',
        password: '',
        confirmPassword: '',
    });

    const [captcha, setCaptcha] = useState('');
    const [error, setError] = useState('');
    const [isFormShow, setIsFormShow] = useState(false);
    const [isLoginForm, setIsLoginForm] = useState(true);

    const dispatch = useAppDispatch();
    const [captchaRandom, setCaptchaRandom] = useState(Math.random());

    const [isMounted, setIsMounted] = useState(false);

    useEffect(() => {
        setIsMounted(true);
        return () => {
            setIsMounted(false);
        };
    }, []);

    if (!isMounted) return null;

    const submitLoginForm = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        if (!loginForm.userEmail || !loginForm.password) {
            toast('请填写所有必填字段', {type: 'error'});
            setError('请填写所有必填字段');
            return;
        }
        userLogin(loginForm, captcha).then((res) => {
            if (!res) {
                toast('登录失败，请检查用户名密码或验证码', {type: 'error'});
                setError('登录失败，请检查用户名密码或验证码');
                setCaptchaRandom(Math.random());
            } else {
                dispatch({type: 'user/initUserInfo', payload: res as UserInfo});
                dispatch({type: 'user/changeLoginStatus', payload: true});
                onClose();
            }
        });
    };

    const submitRegisterForm = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        if (!registerForm.userEmail || !registerForm.password || !registerForm.confirmPassword) {
            toast('请填写所有必填字段', {type: 'error'});
            setError('请填写所有必填字段');
            return;
        }
        if (registerForm.password !== registerForm.confirmPassword) {
            toast('两次输入的密码不一致', {type: 'error'});
            setError('两次输入的密码不一致');
            return;
        }
        userRegister(registerForm, captcha).then((res) => {
            if (!res) {
                toast('注册失败，可能是邮箱已被注册或验证码错误', {type: 'error'});
                setError('注册失败，可能是邮箱已被注册或验证码错误');
                setCaptchaRandom(Math.random());
            } else {
                toast('注册成功，请登录', {type: 'success'});
                setIsLoginForm(true);
                setError('注册成功，请登录');
            }
        });
    };

    const toggleForm = (e: React.MouseEvent<HTMLButtonElement>) => {
        e.preventDefault();
        setIsLoginForm(!isLoginForm);
        setError('');
    };

    return (
        <AnimatePresence>
            {isOpen && (
                <div className={styles.modalOverlay}>
                    <motion.div
                        initial={{opacity: 0}}
                        animate={{opacity: 1}}
                        exit={{opacity: 0}}
                        transition={{
                            type: "tween",
                            duration: 0.2,
                        }}
                    >
                        <motion.div
                            initial={{scale: 0.8, opacity: 0, y: 100}}
                            animate={{scale: 1, opacity: 1, y: -20}}
                            exit={{scale: 0.8, opacity: 0, y: 100}}
                        >
                            <div className={styles.modalContent}>
                                <div style={{
                                    position: 'absolute',
                                    top: '1rem',
                                    right: '1rem',
                                }}>
                                    <motion.div whileHover={{scale: 1.05}} whileTap={{scale: 0.95}}>
                                        <Button variant="ghost" onClick={onClose}>
                                            <CloseIcon/>
                                        </Button>
                                    </motion.div>
                                </div>
                                <h2 className={clsx(styles.title, varela_round.className)}>
                                        <span
                                            className={noto_sans_sc.className}>{isLoginForm ? '登录到' : '注册'} </span>
                                    Grtsinry43&apos;s Blog 😘
                                </h2>
                                {
                                    !isFormShow && (
                                        <Button onClick={() => setIsFormShow(true)} style={{
                                            borderRadius: '0.375rem',
                                            backgroundColor: 'rgba(var(--primary),0.5)',
                                            color: 'rga(var(--foreground))',
                                        }}>
                                            <MailIcon width={16} height={16}/>
                                            通过邮箱 {isLoginForm ? '登录' : '注册'}
                                        </Button>
                                    )
                                }
                                <Separator className="my-4"/>
                                <AnimatePresence mode="wait">
                                    {isFormShow && (
                                        <motion.div
                                            key={isLoginForm ? 'login' : 'register'}
                                            initial={{height: 0, opacity: 0}}
                                            animate={{height: 'auto', opacity: 1}}
                                            exit={{height: 0, opacity: 0}}
                                            transition={{
                                                type: 'spring',
                                                stiffness: 500,
                                                damping: 30,
                                                mass: 1,
                                            }}
                                            style={{
                                                width: '100%',
                                                overflow: 'hidden',
                                            }}
                                        >
                                            {isLoginForm ? (
                                                <form
                                                    style={{
                                                        marginTop: '1rem',
                                                        width: '100%',
                                                    }}
                                                    onSubmit={submitLoginForm}>
                                                    {error && (
                                                        <div className={styles.error}>{error}</div>
                                                    )}
                                                    <div className={styles.formGroup}>
                                                        <div className={styles.label}> 邮箱</div>
                                                        <TextField.Root
                                                            key="login-email"
                                                            style={{
                                                                backgroundColor: 'rgba(var(--foreground), 0.03)',
                                                                flex: '1',
                                                                boxShadow: 'none',
                                                                minHeight: '2rem',
                                                                border: '1px solid rgba(var(--foreground), 0.1)',
                                                                borderBottom: '1px solid rgba(var(--foreground), 0.5)',
                                                                outline: 'none',
                                                                borderRadius: '0.375rem',
                                                            }}
                                                            value={loginForm.userEmail}
                                                            onChange={(e) => setLoginForm({
                                                                ...loginForm,
                                                                userEmail: e.target.value
                                                            })}
                                                        />
                                                    </div>
                                                    <div className={styles.formGroup}>
                                                        <div className={styles.label}> 密码</div>
                                                        <TextField.Root
                                                            key="login-password"
                                                            style={{
                                                                backgroundColor: 'rgba(var(--foreground), 0.03)',
                                                                flex: '1',
                                                                boxShadow: 'none',
                                                                minHeight: '2rem',
                                                                border: '1px solid rgba(var(--foreground), 0.1)',
                                                                borderBottom: '1px solid rgba(var(--foreground), 0.5)',
                                                                outline: 'none',
                                                                borderRadius: '0.375rem',
                                                            }}
                                                            type="password"
                                                            value={loginForm.password}
                                                            onChange={(e) => setLoginForm({
                                                                ...loginForm,
                                                                password: e.target.value
                                                            })}
                                                        />
                                                    </div>
                                                    <div className={styles.formGroup}>
                                                        <div className={styles.label} style={{
                                                            marginRight: '1rem',
                                                        }}> 验证码
                                                        </div>
                                                        <TextField.Root
                                                            key="login-captcha"
                                                            style={{
                                                                backgroundColor: 'rgba(var(--foreground), 0.03)',
                                                                flex: '1',
                                                                boxShadow: 'none',
                                                                minHeight: '2rem',
                                                                border: '1px solid rgba(var(--foreground), 0.1)',
                                                                borderBottom: '1px solid rgba(var(--foreground), 0.5)',
                                                                outline: 'none',
                                                                borderRadius: '0.375rem',
                                                            }}
                                                            value={captcha}
                                                            onChange={(e) => setCaptcha(e.target.value)}
                                                        />
                                                        <img
                                                            src={process.env.NEXT_PUBLIC_BASE_URL + '/captcha' + '?' + captchaRandom}
                                                            alt="captcha"
                                                            className={styles.captcha}
                                                            style={{
                                                                cursor: 'pointer',
                                                                width: '100px',
                                                                marginLeft: '1rem',
                                                                borderRadius: '0.25rem',
                                                            }}
                                                            onClick={() => {
                                                                setCaptchaRandom(Math.random());
                                                            }}
                                                        />
                                                    </div>
                                                    <div className="text-sm text-center">
                                                        忘记密码了？ <Link
                                                        href="/my/reset-password-request"
                                                        color={"#184aff"}
                                                        onClick={() => onClose()}> 点击这里重置密码 </Link>
                                                    </div>
                                                    <div className={styles.formActions}>
                                                        <Button style={{
                                                            marginRight: '1rem',
                                                            backgroundColor: 'rgba(var(--primary),0.5)',
                                                            color: 'rga(var(--foreground))',
                                                        }} type="submit">
                                                            登录
                                                        </Button>
                                                        <Button variant="soft" style={{
                                                            backgroundColor: 'rgba(var(--primary),0.5)',
                                                            color: 'rga(var(--foreground))',
                                                        }} onClick={toggleForm}>
                                                            转到注册
                                                        </Button>
                                                    </div>
                                                </form>
                                            ) : (
                                                <form
                                                    style={{
                                                        marginTop: '1rem',
                                                        width: '100%',
                                                    }}
                                                    onSubmit={submitRegisterForm}>
                                                    {error && (
                                                        <div className={styles.error}>{error}</div>
                                                    )}
                                                    <div className={styles.formGroup}>
                                                        <div className={styles.label}> 昵称</div>
                                                        <TextField.Root
                                                            key="register-nickname"
                                                            style={{
                                                                backgroundColor: 'rgba(var(--foreground), 0.03)',
                                                                flex: '1',
                                                                boxShadow: 'none',
                                                                minHeight: '2rem',
                                                                border: '1px solid rgba(var(--foreground), 0.1)',
                                                                borderBottom: '1px solid rgba(var(--foreground), 0.5)',
                                                                outline: 'none',
                                                                borderRadius: '0.375rem',
                                                            }}
                                                            value={registerForm.nickname}
                                                            onChange={(e) => setRegisterForm({
                                                                ...registerForm,
                                                                nickname: e.target.value
                                                            })}
                                                        />
                                                    </div>
                                                    <div className={styles.formGroup}>
                                                        <div className={styles.label}> 邮箱</div>
                                                        <TextField.Root
                                                            key="register-email"
                                                            style={{
                                                                backgroundColor: 'rgba(var(--foreground), 0.03)',
                                                                flex: '1',
                                                                boxShadow: 'none',
                                                                minHeight: '2rem',
                                                                border: '1px solid rgba(var(--foreground), 0.1)',
                                                                borderBottom: '1px solid rgba(var(--foreground), 0.5)',
                                                                outline: 'none',
                                                                borderRadius: '0.375rem',
                                                            }}
                                                            value={registerForm.userEmail}
                                                            onChange={(e) => setRegisterForm({
                                                                ...registerForm,
                                                                userEmail: e.target.value
                                                            })}
                                                        />
                                                    </div>
                                                    <div className={styles.formGroup}>
                                                        <div className={styles.label}> 密码</div>
                                                        <TextField.Root
                                                            key="register-password"
                                                            style={{
                                                                backgroundColor: 'rgba(var(--foreground), 0.03)',
                                                                flex: '1',
                                                                boxShadow: 'none',
                                                                minHeight: '2rem',
                                                                border: '1px solid rgba(var(--foreground), 0.1)',
                                                                borderBottom: '1px solid rgba(var(--foreground), 0.5)',
                                                                outline: 'none',
                                                                borderRadius: '0.375rem',
                                                            }}
                                                            type="password"
                                                            value={registerForm.password}
                                                            onChange={(e) => setRegisterForm({
                                                                ...registerForm,
                                                                password: e.target.value
                                                            })}
                                                        />
                                                    </div>
                                                    <div className={styles.formGroup}>
                                                        <div className={styles.label} style={{
                                                            marginRight: '1rem',
                                                        }}> 确认密码
                                                        </div>
                                                        <TextField.Root
                                                            key="register-confirm-password"
                                                            style={{
                                                                backgroundColor: 'rgba(var(--foreground), 0.03)',
                                                                flex: '1',
                                                                boxShadow: 'none',
                                                                minHeight: '2rem',
                                                                border: '1px solid rgba(var(--foreground), 0.1)',
                                                                borderBottom: '1px solid rgba(var(--foreground), 0.5)',
                                                                outline: 'none',
                                                                borderRadius: '0.375rem',
                                                            }}
                                                            type="password"
                                                            value={registerForm.confirmPassword}
                                                            onChange={(e) => setRegisterForm({
                                                                ...registerForm,
                                                                confirmPassword: e.target.value
                                                            })}
                                                        />
                                                    </div>
                                                    <div className={styles.formGroup}>
                                                        <div className={styles.label}> 验证码</div>
                                                        <TextField.Root
                                                            key="login-captcha"
                                                            style={{
                                                                backgroundColor: 'rgba(var(--foreground), 0.03)',
                                                                flex: '1',
                                                                boxShadow: 'none',
                                                                minHeight: '2rem',
                                                                border: '1px solid rgba(var(--foreground), 0.1)',
                                                                borderBottom: '1px solid rgba(var(--foreground), 0.5)',
                                                                outline: 'none',
                                                                borderRadius: '0.375rem',
                                                            }}
                                                            value={captcha}
                                                            onChange={(e) => setCaptcha(e.target.value)}
                                                        />
                                                        <img
                                                            src={process.env.NEXT_PUBLIC_BASE_URL + '/captcha' + '?' + captchaRandom}
                                                            alt="captcha"
                                                            className={styles.captcha}
                                                            style={{
                                                                cursor: 'pointer',
                                                                width: '100px',
                                                                marginLeft: '1rem',
                                                                borderRadius: '0.25rem',
                                                            }}
                                                            onClick={() => {
                                                                setCaptchaRandom(Math.random());
                                                            }}
                                                        />
                                                    </div>
                                                    <div className={styles.formActions}>
                                                        <Button style={{marginRight: '1rem'}} type="submit">
                                                            注册
                                                        </Button>
                                                        <Button variant="soft" onClick={toggleForm}>
                                                            返回登录
                                                        </Button>
                                                    </div>
                                                </form>
                                            )}
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                                {isFormShow && (
                                    <Button variant="ghost"
                                            style={{
                                                color: 'rga(var(--foreground))',
                                            }}
                                            onClick={() => setIsFormShow(false)}> 返回使用快捷登录 </Button>
                                )}
                                {!isFormShow && (
                                    <>
                                        <div style={{
                                            marginTop: '1rem',
                                            marginBottom: '1rem',
                                            display: 'flex',
                                            alignItems: 'center',
                                            justifyContent: 'center',
                                            fontSize: '0.75rem',
                                        }}> 通过社交账号登录
                                        </div>
                                        <div>
                                            <Tooltip content="使用 GitHub 登录">
                                                <IconButton radius="full" style={{
                                                    backgroundColor: 'rgba(var(--background),0.5)',
                                                    border: '1px solid rgba(var(--foreground), 0.1)',
                                                    color: 'rga(var(--foreground))',
                                                    marginRight: '0.5rem',
                                                }} className={styles.icon} onClick={() => {
                                                    location.href = `${process.env.NEXT_PUBLIC_BASE_URL}/api/v1/oauth2/authorization/github?redirect_uri= ${encodeURIComponent(location.href)}`;
                                                }}>
                                                    <GitHubLogoIcon/>
                                                </IconButton>
                                            </Tooltip>
                                            <Tooltip content="使用 Google 登录">
                                                <IconButton radius="full" style={{
                                                    backgroundColor: 'rgba(var(--background),0.5)',
                                                    border: '1px solid rgba(var(--foreground), 0.1)',
                                                    color: 'rga(var(--foreground))',
                                                }} className={styles.icon} onClick={() => {
                                                    location.href = `${process.env.NEXT_PUBLIC_BASE_URL}/api/v1/oauth2/authorization/google?redirect_uri=${encodeURIComponent(location.href)}`;
                                                }}>
                                                    <FaGoogle/>
                                                </IconButton>
                                            </Tooltip>
                                            {/*<Tooltip content={'使用 Microsoft 登录'}>*/}
                                            {/*    <IconButton radius="full" className={styles.icon}>*/}
                                            {/*        <BiLogoMicrosoft/>*/}
                                            {/*    </IconButton>*/}
                                            {/*</Tooltip>*/}
                                            {/*<Tooltip content={'使用 Apple 登录'}>*/}
                                            {/*    <IconButton radius="full" className={styles.icon}>*/}
                                            {/*        <FaApple/>*/}
                                            {/*    </IconButton>*/}
                                            {/*</Tooltip>*/}
                                            {/*<Tooltip content={'使用微信登录'}>*/}
                                            {/*    <IconButton radius="full" className={styles.icon}>*/}
                                            {/*        <IoLogoWechat/>*/}
                                            {/*    </IconButton>*/}
                                            {/*</Tooltip>*/}
                                        </div>
                                    </>
                                )}
                            </div>
                        </motion.div>
                    </motion.div>
                </div>
            )}
        </AnimatePresence>
    );
};

export default LoginModal;
