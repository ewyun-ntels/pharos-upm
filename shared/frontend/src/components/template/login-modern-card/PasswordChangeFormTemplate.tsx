import React, {useState, useEffect} from 'react';
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
import {TriangleAlert, CheckCircle, AlertCircle} from 'lucide-react';
import {cn} from '../../../lib';
import type {PasswordChangeFormTemplateProps} from './types';
import {validatePasswordPolicy} from '../../../utils/validatePasswordPolicy';

/**
 * 모던 카드 스타일 비밀번호 변경 폼 템플릿
 */
export function PasswordChangeFormTemplate({
                                               onSubmit,
                                               onBack,
                                               errorMessage,
                                               guideMessage,
                                               policyMessage,
                                               passwordPolicy,
                                               isLoading = false,
                                               isExpired = false,
                                               labels = {},
                                           }: PasswordChangeFormTemplateProps) {
    const [currentPassword, setCurrentPassword] = useState('');
    const [newPassword, setNewPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [error, setError] = useState(errorMessage || '');
    const [success, setSuccess] = useState(false);

    // 기본 레이블
    const {
        title = 'Change Password',
        currentPassword: currentPasswordLabel = 'Current Password',
        newPassword: newPasswordLabel = 'New Password',
        confirmPassword: confirmPasswordLabel = 'Confirm Password',
        submit = 'Change Password',
        submitting = 'Changing...',
        passwordMismatch = 'Passwords do not match',
        backToLogin = 'Back to Login',
    } = labels;

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setSuccess(false);

        // 비밀번호 확인
        if (newPassword !== confirmPassword) {
            setError(passwordMismatch);
            return;
        }

        const policyError = validatePasswordPolicy(passwordPolicy, newPassword);
        if (policyError) {
            setError(policyError);
            return;
        }

        try {
            await onSubmit(newPassword, currentPassword);
            setSuccess(true);
            setCurrentPassword('');
            setNewPassword('');
            setConfirmPassword('');
        } catch (err) {
            console.error('Password change failed:', err);
        }
    };

    // 성공 후 자동 뒤로가기 (cleanup으로 메모리 누수 방지)
    useEffect(() => {
        if (!success) return;
        const timer = setTimeout(() => {
            if (onBack) onBack();
        }, 2000);
        return () => clearTimeout(timer);
    }, [success, onBack]);

    const handleBack = () => {
        if (onBack) {
            onBack();
        }
    };

    return (
        <div
            className="shadow-[0_0_0_1px_rgba(0,0,0,0.04),0_0_24px_rgba(0,0,0,0.10),0_0_60px_rgba(0,0,0,0.12)] rounded-2xl">
            <Card
                className="w-full min-w-[380px] rounded-2xl border border-border/60 bg-card shadow-[0_8px_32px_rgba(0,0,0,0.10)] dark:shadow-[0_8px_32px_rgba(0,0,0,0.40)]"
            >
                <CardHeader className="text-center space-y-1.5 pt-9 pb-5 px-9">
                    <CardTitle className="text-[26px] font-semibold tracking-tight">{title}</CardTitle>
                    <CardDescription className="text-sm text-muted-foreground">Update your account
                        password</CardDescription>
                </CardHeader>

                <CardContent className="px-9 pb-9">
                    {guideMessage && (
                        <Alert
                            className="text-yellow-700 dark:text-yellow-400 border-yellow-200 dark:border-yellow-800/50 bg-yellow-50 dark:bg-yellow-900/20 mb-6">
                            <AlertCircle className="h-4 w-4 mr-2 shrink-0"/>
                            <AlertTitle className="line-clamp-none whitespace-normal break-words">
                                {guideMessage}
                            </AlertTitle>
                        </Alert>
                    )}

                    {error && (
                        <Alert
                            className="text-red-700 dark:text-red-400 border-red-200 dark:border-red-800/50 bg-red-50 dark:bg-red-900/20 mb-6">
                            <TriangleAlert className="h-4 w-4 mr-2 shrink-0"/>
                            <AlertTitle className="line-clamp-none whitespace-normal break-words">
                                {error}
                            </AlertTitle>
                        </Alert>
                    )}

                    {success && (
                        <Alert
                            className="text-green-700 dark:text-green-400 border-green-200 dark:border-green-800/50 bg-green-50 dark:bg-green-900/20 mb-6">
                            <CheckCircle className="h-4 w-4 mr-2 shrink-0"/>
                            <AlertTitle>Password changed successfully!</AlertTitle>
                        </Alert>
                    )}

                    <form onSubmit={handleSubmit} className="space-y-5">
                        {/* Current Password (조건부 표시) */}
                        {!isExpired && (
                            <div className="space-y-2">
                                <Label htmlFor="current-password" className="text-sm font-medium">
                                    {currentPasswordLabel}
                                </Label>
                                <Input
                                    id="current-password"
                                    type="password"
                                    placeholder="••••••••"
                                    value={currentPassword}
                                    disabled={isLoading}
                                    onChange={(e) => {
                                        setCurrentPassword(e.target.value);
                                        if (error) setError('');
                                    }}
                                    className="h-10 border border-input rounded-lg focus-visible:ring-0 focus:border-foreground transition-colors"
                                    required
                                />
                            </div>
                        )}

                        {/* New Password */}
                        <div className="space-y-2">
                            <Label htmlFor="new-password" className="text-sm font-medium">
                                {newPasswordLabel}
                            </Label>
                            <Input
                                id="new-password"
                                type="password"
                                placeholder="••••••••"
                                value={newPassword}
                                disabled={isLoading || success}
                                onChange={(e) => {
                                    setNewPassword(e.target.value);
                                    if (error) setError('');
                                }}
                                className="h-11 border border-input rounded-lg focus-visible:ring-0 focus:border-foreground transition-colors"
                                required
                            />
                            {policyMessage && (
                                <p className="text-xs text-muted-foreground">
                                    {policyMessage}
                                </p>
                            )}
                        </div>

                        {/* Confirm Password */}
                        <div className="space-y-2">
                            <Label htmlFor="confirm-password" className="text-sm font-medium">
                                {confirmPasswordLabel}
                            </Label>
                            <Input
                                id="confirm-password"
                                type="password"
                                placeholder="••••••••"
                                value={confirmPassword}
                                disabled={isLoading || success}
                                onChange={(e) => {
                                    setConfirmPassword(e.target.value);
                                    if (error) setError('');
                                }}
                                className="h-11 border border-input rounded-lg focus-visible:ring-0 focus:border-foreground transition-colors"
                                required
                            />
                        </div>

                        {/* Submit Button */}
                        <Button
                            type="submit"
                            disabled={isLoading || success}
                            className={cn(
                                'w-full h-10 font-medium rounded-lg transition-all duration-200',
                                (isLoading || success) && 'opacity-70 cursor-not-allowed'
                            )}
                        >
                            {isLoading ? submitting : success ? 'Success!' : submit}
                        </Button>

                        {/* Back Button */}
                        {onBack && (
                            <Button
                                type="button"
                                variant="outline"
                                disabled={isLoading || success}
                                className="w-full h-10 font-medium rounded-lg"
                                onClick={handleBack}
                            >
                                {backToLogin}
                            </Button>
                        )}
                    </form>
                </CardContent>
            </Card>
        </div>
    );
}
