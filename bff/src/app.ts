import express from 'express';
import dotenv from 'dotenv';
import connectDB from './database';
import userBehaviorRoutes from './routes/userBehaviorRoutes';

dotenv.config();

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

app.use('/api', userBehaviorRoutes);

connectDB().then(r => {
  console.log('MongoDB connected');
});

app.listen(PORT, () => {
  console.log(`Server is running on port ${PORT}`);
});