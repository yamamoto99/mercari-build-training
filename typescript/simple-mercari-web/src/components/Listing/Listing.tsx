import React, { useState } from 'react';

const server = process.env.REACT_APP_API_URL || 'http://127.0.0.1:9000';

interface Prop {
  onListingCompleted?: () => void;
}

type formDataType = {
  name: string,
  category: string,
  image: string | File,
}

export const Listing: React.FC<Prop> = (props) => {
  const { onListingCompleted } = props;
  const initialState = {
    name: "",
    category: "",
    image: "",
  };
  const [values, setValues] = useState<formDataType>(initialState);

  const onValueChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setValues({
      ...values, [event.target.name]: event.target.value,
    })
  };
  const onFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setValues({
      ...values, [event.target.name]: event.target.files![0],
    })
  };
  const onSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const data = new FormData()
    data.append('name', values.name)
    data.append('category', values.category)
    data.append('image', values.image)

    fetch(server.concat('/items'), {
      method: 'POST',
      mode: 'cors',
      body: data,
    })
      .then(response => {
        console.log('POST status:', response.statusText);
        onListingCompleted && onListingCompleted();
      })
      .catch((error) => {
        console.error('POST error:', error);
      })
  };
  return (
    <div className='Listing'>
      <form onSubmit={onSubmit}>
        <div className={"bg-gray-700 p-10 m-20 rounded-lg flex flex-col"}>
          <h2 className={"text-white text-2xl underline decoration-white text-center mb-6"}>Add Item</h2>
          <input type='text' name='name' id='name' placeholder='item name' className='mb-6 bg-gray-100 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 w-20/100 p-2 placeholder:text-slate-500' onChange={onValueChange} required />
          <input type='text' name='category' id='category' placeholder='category' className='mb-6 bg-gray-100 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 w-20/100 p-2 placeholder:text-slate-500' onChange={onValueChange} />
          <input type='file' name='image' id='image' className="mb-6 w-full text-sm  file:py-2 file:px-4 file:border-0 file:text-sm file:bg-blue-600 file:text-white hover:file:bg-blue-700 border border-gray-300 rounded-lg text-gray-900 bg-gray-100" onChange={onFileChange} required />
          <button type='submit' className="bg-blue-600 hover:bg-blue-700 text-white py-2 px-4 rounded">List this item</button>
        </div>
      </form>
    </div>
  );
}
