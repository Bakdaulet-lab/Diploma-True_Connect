'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useForm } from 'react-hook-form';
import { Eye, EyeOff, Phone, Lock } from 'lucide-react';
import { useRegister } from '@/hooks/api';
import { Spinner } from '@/components/ui/common';
import toast from 'react-hot-toast';

interface RegisterForm {
  phone: string;
  password: string;
  confirmPassword: string;
}

export default function RegisterPage() {
  const [showPassword, setShowPassword] = useState(false);
  const registerMutation = useRegister();
  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<RegisterForm>();

  const password = watch('password');

  const onSubmit = (data: RegisterForm) => {
    const phone = data.phone.startsWith('+7') ? data.phone : `+7${data.phone}`;
    registerMutation.mutate(
      { phone, password: data.password },
      { onError: (err: any) => toast.error(err.response?.data?.message || 'Registration failed') }
    );
  };

  return (
    <div>
      <div className="mb-8 lg:hidden flex justify-center">
        <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-primary-500 text-white">
          <svg viewBox="0 0 24 24" className="h-8 w-8" fill="currentColor">
            <path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z" />
          </svg>
        </div>
      </div>

      <h2 className="text-2xl font-bold text-gray-900">Create account</h2>
      <p className="mt-2 text-sm text-gray-500">Join TrueConnect and start building trust</p>

      <form onSubmit={handleSubmit(onSubmit)} className="mt-8 space-y-5">
        <div>
          <label className="mb-1.5 block text-sm font-medium text-gray-700">Phone number</label>
          <div className="relative">
            <Phone className="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-gray-400" />
            <input
              {...register('phone', {
                required: 'Phone number is required',
                pattern: { value: /^\+?7?\d{10}$/, message: 'Enter valid KZ phone (+7XXXXXXXXXX)' },
              })}
              type="tel"
              placeholder="+7 (XXX) XXX-XX-XX"
              className="input-field pl-10"
            />
          </div>
          {errors.phone && <p className="mt-1 text-xs text-red-500">{errors.phone.message}</p>}
        </div>

        <div>
          <label className="mb-1.5 block text-sm font-medium text-gray-700">Password</label>
          <div className="relative">
            <Lock className="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-gray-400" />
            <input
              {...register('password', {
                required: 'Password is required',
                minLength: { value: 8, message: 'At least 8 characters' },
              })}
              type={showPassword ? 'text' : 'password'}
              placeholder="Create a password"
              className="input-field pl-10 pr-10"
            />
            <button
              type="button"
              onClick={() => setShowPassword(!showPassword)}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
            >
              {showPassword ? <EyeOff className="h-5 w-5" /> : <Eye className="h-5 w-5" />}
            </button>
          </div>
          {errors.password && <p className="mt-1 text-xs text-red-500">{errors.password.message}</p>}
        </div>

        <div>
          <label className="mb-1.5 block text-sm font-medium text-gray-700">Confirm password</label>
          <div className="relative">
            <Lock className="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-gray-400" />
            <input
              {...register('confirmPassword', {
                required: 'Please confirm your password',
                validate: (val) => val === password || 'Passwords do not match',
              })}
              type={showPassword ? 'text' : 'password'}
              placeholder="Confirm your password"
              className="input-field pl-10"
            />
          </div>
          {errors.confirmPassword && (
            <p className="mt-1 text-xs text-red-500">{errors.confirmPassword.message}</p>
          )}
        </div>

        <button type="submit" disabled={registerMutation.isPending} className="btn-primary w-full">
          {registerMutation.isPending ? <Spinner className="h-5 w-5 text-white" /> : 'Create account'}
        </button>
      </form>

      <p className="mt-6 text-center text-sm text-gray-500">
        Already have an account?{' '}
        <Link href="/login" className="font-semibold text-primary-500 hover:text-primary-600">
          Sign in
        </Link>
      </p>
    </div>
  );
}
