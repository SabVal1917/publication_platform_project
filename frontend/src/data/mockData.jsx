export const mockArticles = [
  {
    id: 1,
    title: "Algorithm usages in React",
    description: "Yet another React guide.",
    tags: ["node", "react", "algorithms"],
    likes: "589",
    date: "Oct 1, 2025, 4:09:15 PM",
    readTime: 10,
    content: `
      <p>A concise walkthrough covering pitfalls, best practices, and performance tricks for real-world applications.</p>
      <h2>Example</h2>
      <pre><code>function showMessage() {
  alert( 'Всем привет!' );
}</code></pre>
      <p>Leveraging algorithms in React isn't just about sorting lists. It's about optimizing renders, managing state transitions efficiently, and creating a more responsive user experience. For instance, memoization is a direct application of caching algorithmic principles.</p>
    `,
  },
  {
    id: 2,
    title: "Designing Efficient Algorithms for Interviews #1",
    description:
      "A concise walkthrough covering pitfalls, best practices, and performance tricks for real-world…",
    tags: ["react", "javascript"],
    likes: "541",
    date: "Oct 1, 2025, 10:09:15 PM",
    readTime: 6,
    content: `
      <p>This is the main content for the article "Designing Efficient Algorithms". It delves into Big O notation, common data structures, and how to approach problem-solving during a technical interview.</p>
      <h2>Pitfalls</h2>
      <p>One common pitfall is premature optimization. Always write clear, correct code first, then identify bottlenecks and optimize them.</p>
    `,
  },
  {
    id: 3,
    title: "State Management in Modern JS",
    description:
      "Exploring Zustand, Redux Toolkit, and native Context API for scalable applications.",
    tags: ["javascript", "react", "state-management"],
    likes: "498",
    date: "Sep 30, 2025, 8:00:00 AM",
    readTime: 15,
    content: `
      <p>Zustand offers a minimal and straightforward approach, while Redux Toolkit provides a robust, opinionated structure for large-scale apps. The choice depends on your project's complexity.</p>
    `,
  },
];

export const mockUser = {
  name: "Maximilian Khomyakov",
  username: "@maxkhomyakov",
  avatar: "https://placehold.co/112x112/895AF6/FFFFFF?text=M", // Заглушка для аватара
  bio: "Fullstack junior.",
  location: "Moscow, Russia",
  website: "example.com",
  joined: "June 2022",
  stats: {
    articles: 2,
    followers: "1,240",
    following: "180",
  },
};

// Добавим еще пару статей, чтобы было что отображать в профиле
export const mockUserArticles = [
  {
    id: 4,
    title: "Understanding Closures in JS",
    description:
      "A concise overview of the topic with practical examples and recommendations for real…",
    tags: ["react"],
    likes: "74",
    date: "Sep 10, 2025, 10:36:11 PM",
    readTime: 8,
  },
  {
    id: 5,
    title: "Algorithms for Everyday Devs",
    description:
      "A concise overview of the topic with practical examples and recommendations for real…",
    tags: ["C++", "testing", "performance"],
    likes: "5",
    date: "Aug 27, 2025, 10:36:11 PM",
    readTime: 10,
  },
];
