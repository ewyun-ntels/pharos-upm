import React, {useState} from 'react';
import {
    Button,
    Input,
    Label,
    Alert,
    AlertTitle,
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle
} from '../../ui';
import {TriangleAlert} from 'lucide-react';
import {cn} from '../../../lib';
import type {LoginFormTemplateProps} from './types';

/**
 * 모던 카드 스타일 로그인 폼 템플릿
 *
 * 카드 기반의 모던 디자인으로 사용성이 좋은 로그인 폼을 제공합니다.
 * 이메일/패스워드 입력과 에러 메시지 표시를 지원합니다.
 */
export function LoginFormTemplate({
                                      logo,
                                      onSubmit,
                                      errorMessage,
                                      isLoading = false,
                                      labels = {},
                                  }: LoginFormTemplateProps) {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState(errorMessage || '');

    // 기본 레이블
    const {
        title = 'Sign in',
        subtitle = 'Enter your credentials to access your account',
        username: usernameLabel = 'Email or Username',
        password: passwordLabel = 'Password',
        signIn = 'Sign in',
        signingIn = 'Signing in...',
    } = labels;

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');

        try {
            await onSubmit(username, password);
        } catch (err) {
            console.error('Login failed:', err);
        }
    };

    React.useEffect(() => {
        if (errorMessage) {
            setError(errorMessage);
        }
    }, [errorMessage]);

    return (
        <div
            className="shadow-[0_0_0_1px_rgba(0,0,0,0.04),0_0_24px_rgba(0,0,0,0.10),0_0_60px_rgba(0,0,0,0.12)] rounded-2xl">
            <Card
                className="w-full rounded-2xl border border-border/60 bg-card shadow-[0_8px_32px_rgba(0,0,0,0.10)] dark:shadow-[0_8px_32px_rgba(0,0,0,0.40)]"
            >
                <CardHeader className="text-center space-y-1.5 pt-9 pb-5 px-9">
                    {logo && (
                        <div className="mx-auto mb-4">
                            {logo}
                        </div>
                    )}
                    <CardTitle className="text-[26px] font-semibold tracking-tight">{title}</CardTitle>
                    <CardDescription className="text-sm text-muted-foreground">
                        {subtitle}
                    </CardDescription>
                </CardHeader>

                <CardContent className="space-y-5 px-9 pb-9">
                    {error && (
                        <Alert
                            className="text-red-700 dark:text-red-400 border border-red-200 dark:border-red-800/50 bg-red-50 dark:bg-red-900/20">
                            <TriangleAlert className="h-4 w-4 mr-2 shrink-0 mt-0.5"/>
                            <AlertTitle className="line-clamp-none whitespace-normal wrap-break-word">
                                {error}
                            </AlertTitle>
                        </Alert>
                    )}

                    <form onSubmit={handleSubmit} className="space-y-4">
                        {/* Username/Email 필드 */}
                        <div className="space-y-2">
                            <Label htmlFor="username" className="text-sm font-medium">
                                {usernameLabel}
                            </Label>
                            <Input
                                id="username"
                                type="text"
                                placeholder="name@example.com"
                                value={username}
                                disabled={isLoading}
                                onChange={(e) => {
                                    setUsername(e.target.value);
                                    if (error) setError('');
                                }}
                                className="h-10 px-3 border border-input rounded-lg bg-background text-foreground placeholder:text-muted-foreground/60 focus-visible:ring-0 focus:border-foreground transition-colors"
                                required
                            />
                        </div>

                        {/* Password 필드 */}
                        <div className="space-y-2">
                            <Label htmlFor="password" className="text-sm font-medium">
                                {passwordLabel}
                            </Label>
                            <Input
                                id="password"
                                type="password"
                                placeholder="••••••••"
                                value={password}
                                disabled={isLoading}
                                onChange={(e) => {
                                    setPassword(e.target.value);
                                    if (error) setError('');
                                }}
                                className="h-10 px-3 border border-input rounded-lg bg-background text-foreground placeholder:text-muted-foreground/60 focus-visible:ring-0 focus:border-foreground transition-colors"
                                required
                            />
                        </div>

                        {/* Sign in Button */}
                        <Button
                            type="submit"
                            disabled={isLoading}
                            className={cn(
                                'w-full h-10 font-semibold rounded-lg transition-all duration-200 mt-2',
                                isLoading && 'opacity-60 cursor-not-allowed'
                            )}
                        >
                            {isLoading ? signingIn : signIn}
                        </Button>
                    </form>
                </CardContent>
            </Card>
        </div>
    );
}
