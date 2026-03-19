import { Loader2 } from 'lucide-react';

const sizes = { sm: 'h-4 w-4', md: 'h-6 w-6', lg: 'h-10 w-10' } as const;

interface LoadingSpinnerProps {
  size?: keyof typeof sizes;
  fullPage?: boolean;
}

export default function LoadingSpinner({ size = 'md', fullPage = false }: LoadingSpinnerProps) {
  const spinner = <Loader2 className={`${sizes[size]} animate-spin text-gray-400`} />;

  if (fullPage) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        {spinner}
      </div>
    );
  }
  return spinner;
}
