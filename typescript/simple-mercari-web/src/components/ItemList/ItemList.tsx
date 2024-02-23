import React, { useEffect, useState } from 'react';

interface Item {
  id: number;
  name: string;
  category: string;
  image_name: string;
};

const server = process.env.REACT_APP_API_URL || 'http://127.0.0.1:9000';
const placeholderImage = process.env.PUBLIC_URL + '/logo192.png';

interface Prop {
  reload?: boolean;
  onLoadCompleted?: () => void;
}

export const ItemList: React.FC<Prop> = (props) => {
  const { reload = true, onLoadCompleted } = props;
  const [items, setItems] = useState<Item[]>([])
  const fetchItems = () => {
    fetch(server.concat('/items'),
      {
        method: 'GET',
        mode: 'cors',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        },
      })
      .then(response => response.json())
      .then(data => {
        console.log('GET success:', data);
        setItems(data.items);
        onLoadCompleted && onLoadCompleted();
      })
      .catch(error => {
        console.error('GET error:', error)
      })
  }

  useEffect(() => {
    if (reload) {
      fetchItems();
    }
  }, [reload]);

  return (
    <div className="flex flex-wrap m-5">
      {items?.map((item) => {
        return (
          <div className="p-2" key={item.id}>
            <div className='max-w-64 rounded overflow-hidden shadow-lg bg-white'>
              {/* TODO: Task 1: Replace the placeholder image with the item image */}
              <img className="w-64 h-64 border" src={"http://localhost:9000/image/" + item.id}/>
              <div className="px-6 py-4 border">
                <span className="font-bold text-xl mb-2">{item.name}</span>
                <br/>
                <span className="text-gray-700 text-base">{item.category}</span>
              </div>
            </div>
          </div>
        )
      })}
    </div>
  )
};
