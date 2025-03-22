'use client';

import React, {useState} from 'react';
import {AnimatePresence, motion} from 'framer-motion';
import {IconButton, TextField, Tooltip} from '@radix-ui/themes';
import {Button, ConfigProvider} from 'react-vant';
import styles from '@/styles/LoginModalMobile.module.scss';
import {clsx} from 'clsx';
import {varela_round, noto_sans_sc} from '@/app/fonts/font';
import {MailIcon, X} from 'lucide-react';
import {GitHubLogoIcon} from '@radix-ui/react-icons';
import {FaGoogle} from 'react-icons/fa';
import {userLogin, userRegister} from '@/api/user';
import {UserInfo} from '@/redux/userSlice';
import {useAppDispatch} from '@/redux/hooks';
import {Popup} from "react-vant";
import 'react-vant/lib/index.css';
import '@vant/touch-emulator';

const LoginModalMobile = ({isOpen, onClose}: { isOpen: boolean; onClose: () => void }) => {
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

    const submitLoginForm = () => {
        if (!loginForm.userEmail || !loginForm.password) {
            setError('请填写所有必填字段');
            return;
        }
        userLogin(loginForm, captcha).then((res) => {
            if (!res) {
                setError('登录失败，请检查用户名密码或验证码');
                setCaptchaRandom(Math.random());
            } else {

                dispatch({type: 'user/initUserInfo', payload: res as UserInfo});
                dispatch({type: 'user/changeLoginStatus', payload: true});
                onClose();
            }
        });
    };

    const submitRegisterForm = () => {
        if (!registerForm.userEmail || !registerForm.password || !registerForm.confirmPassword) {
            setError('请填写所有必填字段');
            return;
        }
        if (registerForm.password !== registerForm.confirmPassword) {
            setError('两次输入的密码不一致');
            return;
        }
        userRegister(registerForm, captcha).then((res) => {
            if (!res) {
                setError('注册失败，可能是邮箱已被注册或验证码错误');
                setCaptchaRandom(Math.random());
            } else {

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

    const themeVars = {
        '--rv-popup-background-color': 'transparent',
        '--rv-popup-transition': 'transform var(--rv-animation-duration-base)',
        '--rv-popup-round-border-radius': '16px',
    };

    return (
        <div>
            <ConfigProvider themeVars={themeVars}>
                <Popup round visible={isOpen} onClose={onClose} position="bottom" style={{
                    backgroundColor: 'transparent',
                }}>
                    <AnimatePresence>
                        {isOpen && (
                            <div className={styles.modalOverlay}>
                                <motion.div
                                    initial={{opacity: 0}}
                                    animate={{opacity: 1}}
                                    exit={{opacity: 0}}
                                >
                                    <motion.div
                                        initial={{scale: 0.8, opacity: 0}}
                                        animate={{scale: 1, opacity: 1}}
                                        exit={{scale: 0.8, opacity: 0}}
                                    >
                                        <div className={styles.modalContent}>
                                            <div style={{
                                                position: 'absolute',
                                                top: '1rem',
                                                right: '1rem',
                                            }}>
                                                <motion.div whileHover={{scale: 1.05}} whileTap={{scale: 0.95}}>
                                                    <Button onClick={onClose} style={{
                                                        backgroundColor: 'transparent',
                                                        border: 'none',
                                                    }}>
                                                        <X/>
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
                                                        backgroundColor: 'transparent',
                                                        borderRadius: '0.5rem',
                                                        border: '1px solid var(--primary-color)',
                                                        color: 'var(--primary-color)',
                                                    }}>
                                                        <div style={{
                                                            display: 'flex',
                                                        }}>
                                                            <MailIcon size={'16'} className="mr-3"/>
                                                            <span> 通过邮箱 {isLoginForm ? '登录' : '注册'}</span>
                                                        </div>
                                                    </Button>
                                                )
                                            }
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
                                                                }}>
                                                                {error && (
                                                                    <div className={styles.error}>{error}</div>
                                                                )}
                                                                <div className={styles.formGroup}>
                                                                    <div className={styles.label}> 邮箱</div>
                                                                    <TextField.Root
                                                                        key="login-email"
                                                                        className={styles.textField}
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
                                                                        className={styles.textField}
                                                                        type="password"
                                                                        value={loginForm.password}
                                                                        onChange={(e) => setLoginForm({
                                                                            ...loginForm,
                                                                            password: e.target.value
                                                                        })}
                                                                    />
                                                                </div>
                                                                <div className={styles.formGroup}>
                                                                    <div className={styles.label}> 验证码</div>
                                                                    <TextField.Root
                                                                        key="login-captcha"
                                                                        className={styles.textField}
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
                                                                    <Button style={{
                                                                        marginRight: '1rem',
                                                                        backgroundColor: 'var(--primary-color)',
                                                                        color: 'var(--background-color)',
                                                                        border: 'none',
                                                                    }}
                                                                            onClick={() => submitLoginForm()}
                                                                    >
                                                                        登录
                                                                    </Button>
                                                                    <Button style={{
                                                                        backgroundColor: 'transparent',
                                                                        border: '1px solid var(--primary-color)',
                                                                        color: 'var(--primary-color)',
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
                                                                        className={styles.textField}
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
                                                                        className={styles.textField}
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
                                                                        className={styles.textField}
                                                                        type="password"
                                                                        value={registerForm.password}
                                                                        onChange={(e) => setRegisterForm({
                                                                            ...registerForm,
                                                                            password: e.target.value
                                                                        })}
                                                                    />
                                                                </div>
                                                                <div className={styles.formGroup}>
                                                                    <div className={styles.label}> 确认密码</div>
                                                                    <TextField.Root
                                                                        key="register-confirm-password"
                                                                        className={styles.textField}
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
                                                                        className={styles.textField}
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
                                                                    <Button style={{
                                                                        marginRight: '1rem',
                                                                        backgroundColor: 'var(--primary-color)',
                                                                        color: 'var(--background-color)',
                                                                        border: 'none',
                                                                    }} onClick={
                                                                        () => submitRegisterForm()
                                                                    }>
                                                                        注册
                                                                    </Button>
                                                                    <Button style={{
                                                                        backgroundColor: 'transparent',
                                                                        border: '1px solid var(--primary-color)',
                                                                        color: 'var(--primary-color)',
                                                                    }} onClick={toggleForm}>
                                                                        返回登录
                                                                    </Button>
                                                                </div>
                                                            </form>
                                                        )}
                                                    </motion.div>
                                                )}
                                            </AnimatePresence>
                                            {isFormShow && (
                                                <Button style={{
                                                    backgroundColor: 'transparent',
                                                    border: 'none',
                                                    color: 'var(--primary-color)',
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
                                                            <IconButton radius="full" className={styles.icon}
                                                                        onClick={() => {
                                                                            location.href = `${process.env.NEXT_PUBLIC_BASE_URL}/api/v1/oauth2/authorization/github?redirect_uri= ${encodeURIComponent(location.href)}`;
                                                                        }}>
                                                                <GitHubLogoIcon/>
                                                            </IconButton>
                                                        </Tooltip>
                                                        <Tooltip content="使用 Google 登录">
                                                            <IconButton radius="full" className={styles.icon}
                                                                        onClick={() => {
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
                </Popup>
            </ConfigProvider>
        </div>
    );
};

export default LoginModalMobile;
