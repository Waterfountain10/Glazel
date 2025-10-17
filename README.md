i made Glazel cuz my C++ projects like my Game Boy emulator take forever to build sometimes, like 40s-50s cold for complex c++ projects, and it adds up when I keep changing stuff. I switch between arm mac, windows, (and my friends whom i collab often with use linux) so it’s annoying rebuilding the same files everywhere. 

tools like bazel or distcc didn’t really fit my setup so I built my own thing that shares build results across devices, cuts build time by a *lot* (~5s build time), and just makes my workflow way faster.

-- W
