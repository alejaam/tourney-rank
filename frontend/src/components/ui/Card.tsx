import { cn } from '../../lib/utils';

interface CardProps {
    children: React.ReactNode;
    className?: string;
}

export const Card = ({ children, className }: CardProps) => {
    return (
        <div className={cn('bg-gray-800 rounded-xl shadow-lg p-6', className)}>
            {children}
        </div>
    );
};

export const CardHeader = ({ children, className }: CardProps) => {
    return <div className={cn('mb-4', className)}>{children}</div>;
};

export const CardTitle = ({ children, className }: CardProps) => {
    return <h2 className={cn('text-2xl font-bold text-white', className)}>{children}</h2>;
};

export const CardDescription = ({ children, className }: CardProps) => {
    return <p className={cn('text-gray-400 mt-1', className)}>{children}</p>;
};

export const CardContent = ({ children, className }: CardProps) => {
    return <div className={cn('', className)}>{children}</div>;
};

export const CardFooter = ({ children, className }: CardProps) => {
    return <div className={cn('mt-6 pt-4 border-t border-gray-700', className)}>{children}</div>;
};
