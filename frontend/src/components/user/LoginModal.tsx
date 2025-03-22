'use client';

import React, { useEffect, useState } from 'react';
import { AnimatePresence, motion } from 'framer-motion';
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Separator } from '@/components/ui/separator';
import { X, MailIcon, LucideGithub } from 'lucide-react';
import { FaGoogle } from 'react-icons/fa';
import { userLogin, userRegister } from '@/api/user';
import { UserInfo } from '@/redux/userSlice';
import { useAppDispatch } from '@/redux/hooks';
import Link from "next/link";
import { toast } from "react-toastify";

const LoginModal = ({ isOpen, onClose }: { isOpen: boolean; onClose: () => void }) => {
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
            toast('请填写所有必填字段', { type: 'error' });
            setError('请填写所有必填字段');
            return;
        }
        userLogin(loginForm, captcha).then((res) => {
            if (!res) {
                toast('登录失败，请检查用户名密码或验证码', { type: 'error' });
                setError('登录失败，请检查用户名密码或验证码');
                setCaptchaRandom(Math.random());
            } else {
                dispatch({ type: 'user/initUserInfo', payload: res as UserInfo });
                dispatch({ type: 'user/changeLoginStatus', payload: true });
                onClose();
            }
        });
    };

    const submitRegisterForm = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        if (!registerForm.userEmail || !registerForm.password || !registerForm.confirmPassword) {
            toast('请填写所有必填字段', { type: 'error' });
            setError('请填写所有必填字段');
            return;
        }
        if (registerForm.password !== registerForm.confirmPassword) {
            toast('两次输入的密码不一致', { type: 'error' });
            setError('两次输入的密码不一致');
            return;
        }
        userRegister(registerForm, captcha).then((res) => {
            if (!res) {
                toast('注册失败，可能是邮箱已被注册或验证码错误', { type: 'error' });
                setError('注册失败，可能是邮箱已被注册或验证码错误');
                setCaptchaRandom(Math.random());
            } else {
                toast('注册成功，请登录', { type: 'success' });
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
                <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm flex items-center justify-center">
                    <motion.div
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        exit={{ opacity: 0 }}
                        transition={{
                            type: "tween",
                            duration: 0.2,
                        }}
                        className="w-full h-full flex items-center justify-center"
                    >
                        <motion.div
                            initial={{ scale: 0.8, opacity: 0, y: 100 }}
                            animate={{ scale: 1, opacity: 1, y: 0 }}
                            exit={{ scale: 0.8, opacity: 0, y: 100 }}
                            className="relative bg-card rounded-xl shadow-lg border border-border w-full max-w-md mx-4 overflow-hidden"
                        >
                            <div className="absolute top-4 right-4 z-10">
                                <motion.div whileHover={{ scale: 1.05 }} whileTap={{ scale: 0.95 }}>
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        onClick={onClose}
                                        className="h-8 w-8 rounded-full"
                                    >
                                        <X className="h-4 w-4" />
                                    </Button>
                                </motion.div>
                            </div>

                            <div className="p-6 md:p-8">
                                <div className="space-y-6">
                                    <div className="text-center space-y-2">
                                        <h2 className="text-2xl font-bold tracking-tight">
                                            {isLoginForm ? '登录到' : '注册'} Grtsinry43&apos;s Blog 😘
                                        </h2>
                                        <p className="text-sm text-muted-foreground">
                                            {isLoginForm ? '欢迎回来！请登录您的账号' : '创建一个新账号，开始您的旅程'}
                                        </p>
                                    </div>

                                    {!isFormShow && (
                                        <Button
                                            onClick={() => setIsFormShow(true)}
                                            className="w-full"
                                            variant="default"
                                        >
                                            <MailIcon className="mr-2 h-4 w-4" />
                                            通过邮箱 {isLoginForm ? '登录' : '注册'}
                                        </Button>
                                    )}

                                    <AnimatePresence mode="wait">
                                        {isFormShow && (
                                            <motion.div
                                                key={isLoginForm ? 'login' : 'register'}
                                                initial={{ height: 0, opacity: 0 }}
                                                animate={{ height: 'auto', opacity: 1 }}
                                                exit={{ height: 0, opacity: 0 }}
                                                transition={{
                                                    type: 'spring',
                                                    stiffness: 500,
                                                    damping: 30,
                                                    mass: 1,
                                                }}
                                                className="w-full overflow-hidden"
                                            >
                                                {isLoginForm ? (
                                                    <form
                                                        className="space-y-4"
                                                        onSubmit={submitLoginForm}
                                                    >
                                                        {error && (
                                                            <div className="p-3 text-sm bg-destructive/10 border border-destructive/20 text-destructive rounded-md">
                                                                {error}
                                                            </div>
                                                        )}
                                                        <div className="space-y-2">
                                                            <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                                                邮箱
                                                            </label>
                                                            <Input
                                                                type="email"
                                                                placeholder="your@email.com"
                                                                value={loginForm.userEmail}
                                                                onChange={(e) => setLoginForm({
                                                                    ...loginForm,
                                                                    userEmail: e.target.value
                                                                })}
                                                            />
                                                        </div>
                                                        <div className="space-y-2">
                                                            <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                                                密码
                                                            </label>
                                                            <Input
                                                                type="password"
                                                                placeholder="••••••••"
                                                                value={loginForm.password}
                                                                onChange={(e) => setLoginForm({
                                                                    ...loginForm,
                                                                    password: e.target.value
                                                                })}
                                                            />
                                                        </div>
                                                        <div className="space-y-2">
                                                            <div className="flex items-center justify-between">
                                                                <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                                                    验证码
                                                                </label>
                                                            </div>
                                                            <div className="flex space-x-2">
                                                                <Input
                                                                    placeholder="输入验证码"
                                                                    value={captcha}
                                                                    onChange={(e) => setCaptcha(e.target.value)}
                                                                />
                                                                <div
                                                                    className="flex-shrink-0 h-10 w-24 overflow-hidden rounded-md border cursor-pointer"
                                                                    onClick={() => setCaptchaRandom(Math.random())}
                                                                >
                                                                    <img
                                                                        src={`${process.env.NEXT_PUBLIC_BASE_URL}/captcha?${captchaRandom}`}
                                                                        alt="验证码"
                                                                        className="h-full w-full object-cover"
                                                                    />
                                                                </div>
                                                            </div>
                                                        </div>
                                                        <div className="text-sm text-center">
                                                            <Link
                                                                href="/my/reset-password-request"
                                                                className="text-primary hover:underline"
                                                            >
                                                                忘记密码？点击这里重置
                                                            </Link>
                                                        </div>
                                                        <div className="flex flex-col space-y-2">
                                                            <Button type="submit">
                                                                登录
                                                            </Button>
                                                            <Button
                                                                type="button"
                                                                variant="outline"
                                                                onClick={toggleForm}
                                                            >
                                                                没有账号？注册
                                                            </Button>
                                                        </div>
                                                    </form>
                                                ) : (
                                                    <form
                                                        className="space-y-4"
                                                        onSubmit={submitRegisterForm}
                                                    >
                                                        {error && (
                                                            <div className="p-3 text-sm bg-destructive/10 border border-destructive/20 text-destructive rounded-md">
                                                                {error}
                                                            </div>
                                                        )}
                                                        <div className="space-y-2">
                                                            <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                                                昵称
                                                            </label>
                                                            <Input
                                                                placeholder="您的昵称"
                                                                value={registerForm.nickname}
                                                                onChange={(e) => setRegisterForm({
                                                                    ...registerForm,
                                                                    nickname: e.target.value
                                                                })}
                                                            />
                                                        </div>
                                                        <div className="space-y-2">
                                                            <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                                                邮箱
                                                            </label>
                                                            <Input
                                                                type="email"
                                                                placeholder="your@email.com"
                                                                value={registerForm.userEmail}
                                                                onChange={(e) => setRegisterForm({
                                                                    ...registerForm,
                                                                    userEmail: e.target.value
                                                                })}
                                                            />
                                                        </div>
                                                        <div className="space-y-2">
                                                            <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                                                密码
                                                            </label>
                                                            <Input
                                                                type="password"
                                                                placeholder="••••••••"
                                                                value={registerForm.password}
                                                                onChange={(e) => setRegisterForm({
                                                                    ...registerForm,
                                                                    password: e.target.value
                                                                })}
                                                            />
                                                        </div>
                                                        <div className="space-y-2">
                                                            <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                                                确认密码
                                                            </label>
                                                            <Input
                                                                type="password"
                                                                placeholder="••••••••"
                                                                value={registerForm.confirmPassword}
                                                                onChange={(e) => setRegisterForm({
                                                                    ...registerForm,
                                                                    confirmPassword: e.target.value
                                                                })}
                                                            />
                                                        </div>
                                                        <div className="space-y-2">
                                                            <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                                                                验证码
                                                            </label>
                                                            <div className="flex space-x-2">
                                                                <Input
                                                                    placeholder="输入验证码"
                                                                    value={captcha}
                                                                    onChange={(e) => setCaptcha(e.target.value)}
                                                                />
                                                                <div
                                                                    className="flex-shrink-0 h-10 w-24 overflow-hidden rounded-md border cursor-pointer"
                                                                    onClick={() => setCaptchaRandom(Math.random())}
                                                                >
                                                                    <img
                                                                        src={`${process.env.NEXT_PUBLIC_BASE_URL}/captcha?${captchaRandom}`}
                                                                        alt="验证码"
                                                                        className="h-full w-full object-cover"
                                                                    />
                                                                </div>
                                                            </div>
                                                        </div>
                                                        <div className="flex flex-col space-y-2">
                                                            <Button type="submit">
                                                                注册
                                                            </Button>
                                                            <Button
                                                                type="button"
                                                                variant="outline"
                                                                onClick={toggleForm}
                                                            >
                                                                已有账号？登录
                                                            </Button>
                                                        </div>
                                                    </form>
                                                )}
                                            </motion.div>
                                        )}
                                    </AnimatePresence>

                                    {isFormShow && (
                                        <Button
                                            variant="ghost"
                                            className="w-full text-sm"
                                            onClick={() => setIsFormShow(false)}
                                        >
                                            返回使用快捷登录
                                        </Button>
                                    )}

                                    {!isFormShow && (
                                        <>
                                            <div className="relative">
                                                <div className="absolute inset-0 flex items-center">
                                                    <Separator className="w-full" />
                                                </div>
                                                <div className="relative flex justify-center text-xs uppercase">
                                                    <span className="bg-card px-2 text-muted-foreground">
                                                        或通过社交账号登录
                                                    </span>
                                                </div>
                                            </div>

                                            <div className="grid grid-cols-2 gap-3">
                                                <Button
                                                    variant="outline"
                                                    className="w-full"
                                                    onClick={() => {
                                                        location.href = `${process.env.NEXT_PUBLIC_BASE_URL}/api/v1/oauth2/authorization/github?redirect_uri=${encodeURIComponent(location.href)}`;
                                                    }}
                                                >
                                                    <LucideGithub className="mr-2 h-4 w-4" />
                                                    GitHub
                                                </Button>
                                                <Button
                                                    variant="outline"
                                                    className="w-full"
                                                    onClick={() => {
                                                        location.href = `${process.env.NEXT_PUBLIC_BASE_URL}/api/v1/oauth2/authorization/google?redirect_uri=${encodeURIComponent(location.href)}`;
                                                    }}
                                                >
                                                    <FaGoogle className="mr-2 h-4 w-4" />
                                                    Google
                                                </Button>
                                            </div>
                                        </>
                                    )}
                                </div>
                            </div>
                        </motion.div>
                    </motion.div>
                </div>
            )}
        </AnimatePresence>
    );
};

export default LoginModal;
