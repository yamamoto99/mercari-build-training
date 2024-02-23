import { useState } from 'react';
import { ItemList } from './components/ItemList';
import { Listing } from './components/Listing';

function App() {
    // reload ItemList after Listing complete
    const [reload, setReload] = useState(true);
    return (
    <div className={"flex flex-col"}>
        <h1　className={"text-white text-4xl text-center m-5"}>Product List</h1>
        <div>
            <Listing onListingCompleted={() => setReload(true)} />
        </div>
        <div>
            <ItemList reload={reload} onLoadCompleted={() => setReload(false)} />
        </div>
    </div>
  )
}

export default App;