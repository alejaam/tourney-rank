import { Link, Outlet, useNavigate } from "react-router-dom";
import { Button } from "../components/ui/Button";
import { useAuthStore } from "../store/authStore";

export function MainLayout() {
    const navigate = useNavigate();
    const { user, logout } = useAuthStore();

    const handleLogout = () => {
        logout();
        navigate("/login");
    };

    const isAdmin = user?.role === "admin";

    return (
        <div className="min-h-screen bg-gray-50">
            {/* Navbar */}
            <nav className="bg-white shadow-sm border-b border-gray-200">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                    <div className="flex justify-between h-16">
                        {/* Left side - Logo and main nav */}
                        <div className="flex">
                            <Link
                                to="/dashboard"
                                className="flex items-center px-2 text-xl font-bold text-blue-600 hover:text-blue-700"
                            >
                                TourneyRank
                            </Link>
                            <div className="hidden sm:ml-6 sm:flex sm:space-x-4">
                                <Link
                                    to="/dashboard"
                                    className="inline-flex items-center px-3 py-2 text-sm font-medium text-gray-700 hover:text-gray-900 hover:bg-gray-50 rounded-md transition-colors"
                                >
                                    Dashboard
                                </Link>
                                <Link
                                    to="/tournaments"
                                    className="inline-flex items-center px-3 py-2 text-sm font-medium text-gray-700 hover:text-gray-900 hover:bg-gray-50 rounded-md transition-colors"
                                >
                                    Tournaments
                                </Link>
                                {isAdmin && (
                                    <Link
                                        to="/tournament-admin"
                                        className="inline-flex items-center px-3 py-2 text-sm font-medium text-gray-700 hover:text-gray-900 hover:bg-gray-50 rounded-md transition-colors"
                                    >
                                        Admin Panel
                                    </Link>
                                )}
                            </div>
                        </div>

                        {/* Right side - User menu */}
                        <div className="flex items-center space-x-4">
                            <Link
                                to="/profile"
                                className="text-sm font-medium text-gray-700 hover:text-gray-900"
                            >
                                {user?.username || "Profile"}
                            </Link>
                            <Button
                                onClick={handleLogout}
                                size="sm"
                            >
                                Logout
                            </Button>
                        </div>
                    </div>
                </div>
            </nav>

            {/* Page content */}
            <main>
                <Outlet />
            </main>
        </div>
    );
}
