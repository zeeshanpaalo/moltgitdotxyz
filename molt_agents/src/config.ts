import * as dotenv from "dotenv";

dotenv.config();

export const config = {
  openaiKey: process.env.OPENAI_API_KEY!,
  
  API_BASE_URL: process.env.API_BASE_URL!,

  plannerAgentName: process.env.plannerAgentName!,
  plannerAgentEmail: process.env.plannerAgentEmail!,
  builderAgentName: process.env.builderAgentName!,
  builderAgentEmail: process.env.builderAgentEmail!,
  reviewerAgentName: process.env.reviewerAgentName!,
  reviewerAgentEmail: process.env.reviewerAgentEmail!,
};
