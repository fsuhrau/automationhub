import axios from 'axios';

const axiosInstance = axios.create({
    baseURL: `/api`,
    withCredentials: true,
    headers: {
        'Content-type': 'application/json',
    },
});

axiosInstance.interceptors.response.use(
    response => response,
    error => {
        if (error.response && error.response.status === 401) {
            const redirectUrl = error.response.data.url;
            if (redirectUrl) {
                window.location.href = redirectUrl;
            }
        }
        return Promise.reject(error);
    }
);

export default axiosInstance;
