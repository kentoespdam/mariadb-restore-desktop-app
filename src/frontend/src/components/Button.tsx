import clsx from 'clsx';
import type { ComponentProps } from 'react';

type Variant = 'primary' | 'ghost';
type Size = 'md' | 'icon';

const base =
  'inline-flex items-center justify-center gap-2 rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-slate-400 disabled:pointer-events-none disabled:opacity-50 [&_svg]:size-4 [&_svg]:shrink-0';

const variants: Record<Variant, string> = {
  primary: 'bg-slate-700 text-white hover:bg-slate-600 px-4 h-9',
  ghost: 'hover:bg-slate-800 text-slate-300 h-9 w-9',
};

const sizes: Record<Size, string> = {
  md: '',
  icon: '',
};

export function Button({
  variant = 'primary',
  size = 'md',
  className,
  ...rest
}: ComponentProps<'button'> & { variant?: Variant; size?: Size }) {
  return <button {...rest} className={clsx(base, variants[variant], sizes[size], className)} />;
}
