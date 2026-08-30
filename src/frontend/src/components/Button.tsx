import clsx from 'clsx';
import type { ComponentProps } from 'react';

type Variant = 'primary' | 'secondary' | 'outline' | 'ghost' | 'danger';
type Size = 'sm' | 'md' | 'lg' | 'icon';

const base =
  'inline-flex items-center justify-center gap-2 rounded-lg text-sm font-medium transition-all duration-150 select-none focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sky-500/50 focus-visible:ring-offset-2 focus-visible:ring-offset-slate-950 disabled:pointer-events-none disabled:opacity-40 [&_svg]:size-4 [&_svg]:shrink-0 cursor-pointer disabled:cursor-not-allowed';

const variants: Record<Variant, string> = {
  primary:
    'bg-sky-600 hover:bg-sky-500 active:bg-sky-700 text-white font-semibold shadow-sm shadow-sky-950/50',
  secondary:
    'bg-slate-800 hover:bg-slate-700 active:bg-slate-750 text-slate-100 border border-slate-700/80 shadow-sm',
  outline:
    'border border-slate-700 hover:border-slate-600 bg-transparent hover:bg-slate-800/60 text-slate-300 hover:text-white',
  ghost: 'bg-transparent hover:bg-slate-800/80 active:bg-slate-800 text-slate-300 hover:text-white',
  danger:
    'bg-rose-600 hover:bg-rose-500 active:bg-rose-700 text-white font-medium shadow-sm shadow-rose-950/50',
};

const sizes: Record<Size, string> = {
  sm: 'h-8 px-2.5 text-xs',
  md: 'h-9 px-3.5 text-sm',
  lg: 'h-10 px-4 text-sm',
  icon: 'h-9 w-9 p-0',
};

export function Button({
  variant = 'primary',
  size = 'md',
  className,
  ...rest
}: ComponentProps<'button'> & { variant?: Variant; size?: Size }) {
  return <button {...rest} className={clsx(base, variants[variant], sizes[size], className)} />;
}
