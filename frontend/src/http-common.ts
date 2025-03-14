import axios from 'axios';

export default axios.create({
    baseURL: `/api`,
    withCredentials: true,
    headers: {
        'Content-type': 'application/json',
    },
});
