import toast from "react-hot-toast";

// Re-export toast with project-specific defaults
export const showSuccess = (message: string) => {
  return toast.success(message, {
    duration: 3000,
    position: "top-right",
  });
};

export const showError = (message: string) => {
  return toast.error(message, {
    duration: 4000,
    position: "top-right",
  });
};

export const showInfo = (message: string) => {
  return toast(message, {
    duration: 3000,
    position: "top-right",
    icon: "ℹ️",
  });
};

export const showLoading = (message: string) => {
  return toast.loading(message, {
    position: "top-right",
  });
};

// Re-export the default toast for custom use cases
export { toast };
