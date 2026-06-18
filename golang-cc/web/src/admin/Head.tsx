export default function Head(
  { title, text }: 
  { title: string; text: string }) {
  return (
    <div className="head">
      <h1>{title}</h1>
      <p>{text}</p>
    </div>
  );
}
