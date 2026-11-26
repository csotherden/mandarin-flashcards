import { useEffect, useState } from "react";
import { apiBlob } from "../api";
import type { Flashcard } from "../types";

export function CardImage({ card, className }: { card: Flashcard; className?: string }) {
  const [src, setSrc] = useState(card.image_url);

  useEffect(() => {
    if (card.image_url) {
      setSrc(card.image_url);
      return;
    }
    if (!card.has_local_image) {
      setSrc("");
      return;
    }
    let objectUrl = "";
    let cancel = false;
    apiBlob(`/api/flashcards/${card.id}/image`)
      .then((url) => {
        if (cancel) {
          URL.revokeObjectURL(url);
          return;
        }
        objectUrl = url;
        setSrc(url);
      })
      .catch(() => setSrc(""));
    return () => {
      cancel = true;
      if (objectUrl) URL.revokeObjectURL(objectUrl);
    };
  }, [card.has_local_image, card.id, card.image_url]);

  if (!src) {
    return <div className={`${className ?? ""} grid place-items-center bg-stone-200 text-sm text-stone-500`}>No image</div>;
  }
  return <img src={src} alt="" className={`${className ?? ""} object-cover`} />;
}

export function Attribution({ card }: { card: Flashcard }) {
  const credit = card.unsplash_attribution;
  if (!credit?.photographer_name) return null;
  return (
    <p className="text-xs text-stone-500">
      Photo by{" "}
      <a className="underline" href={credit.photographer_url} target="_blank" rel="noreferrer">
        {credit.photographer_name}
      </a>{" "}
      on{" "}
      <a className="underline" href={credit.photo_url || "https://unsplash.com"} target="_blank" rel="noreferrer">
        Unsplash
      </a>
    </p>
  );
}
