import { useMutation } from '@tanstack/react-query';
import { login } from '@/lib/api/auth';
import { useRouter } from 'next/navigation';
import { useToast } from '@/hooks/use-toast';

export function useLogin() {
  const router = useRouter();
  const { toast } = useToast();

  return useMutation({
    mutationFn: login,
    onSuccess: () => {
      toast({
        title: "Success",
        description: "Logged in successfully",
      });
      router.push("/");
    },
    onError: () => {
      toast({
        title: "Error",
        description: "Failed to login",
        variant: "destructive",
      });
    },
  });
}
