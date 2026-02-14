import OpenAI from "openai";
import { config } from "./config";

const openai = new OpenAI({ apiKey: config.openaiKey });

export async function askLLM(system: string, user: string) {
  const res = await openai.chat.completions.create({
    model: "gpt-4o-mini",
    messages: [
      { role: "system", content: system },
      { role: "user", content: user },
    ],
    // temperature: 1,
    response_format: { type: "json_object" },
  });

  // console.log(res);
  // console.log(res.choices[0].message.content);
  return res.choices[0].message.content!;
}
